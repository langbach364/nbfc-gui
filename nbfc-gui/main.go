package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	defaultMargin  = 20
	defaultSpacing = 10
	windowWidth    = 600
	windowHeight   = 400
)

type NBFCCommand struct {
	Command string      `json:"command"`
	Fan     int         `json:"fan"`
	Speed   interface{} `json:"speed,omitempty"`
}

type NBFCStatus struct {
	PID         int     `json:"pid"`
	Config      string  `json:"config"`
	ReadOnly    bool    `json:"read-only"`
	Temperature float64 `json:"temperature"`
	Fans        []Fan   `json:"fans"`
}

type Fan struct {
	Name         string  `json:"name"`
	AutoMode     bool    `json:"automode"`
	Critical     bool    `json:"critical"`
	CurrentSpeed float64 `json:"current_speed"`
	TargetSpeed  float64 `json:"target_speed"`
	SpeedSteps   int     `json:"speed_steps"`
}

type FanControl struct {
	slider     *gtk.Scale
	valueLabel *gtk.Label
	autoButton *gtk.ToggleButton
	lastValue  float64
	fanIndex   int
}

type StatusLabels struct {
	cpuTemp      *gtk.Label
	cpuFanSpeed  *gtk.Label
	cpuFanTarget *gtk.Label
	cpuFanAuto   *gtk.Label
	gpuFanSpeed  *gtk.Label
	gpuFanTarget *gtk.Label
	gpuFanAuto   *gtk.Label
}

func loadGtkConfig() map[string]string {
	config := make(map[string]string)

	// Đọc settings.ini
	if data, err := os.ReadFile(os.ExpandEnv("$HOME/.config/gtk-3.0/settings.ini")); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Đọc gtkrc-2.0
	if data, err := os.ReadFile(os.ExpandEnv("$HOME/.gtkrc-2.0")); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "gtk-theme-name") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					config["gtk-theme-name"] = strings.Trim(parts[1], "\" ")
				}
			}
		}
	}

	// Đọc icon theme
	if data, err := os.ReadFile(os.ExpandEnv("$HOME/.icons/default/index.theme")); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Inherits=") {
				config["icon-theme"] = strings.TrimPrefix(line, "Inherits=")
			}
		}
	}

	return config
}

func sendCommand(cmd NBFCCommand) error {
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("lỗi tạo JSON: %v", err)
	}

	output, err := exec.Command("bash", "-c",
		fmt.Sprintf("echo '%s' | socat - UNIX-CONNECT:/var/run/nbfc_service.socket", string(jsonData))).CombinedOutput()
	if err != nil {
		return fmt.Errorf("lỗi thực thi: %v", err)
	}
	fmt.Printf("Phản hồi: %s\n", output)
	return nil
}

func createFanControl(label string, fanIndex int) (*gtk.Box, *FanControl) {
	control := &FanControl{
		fanIndex:   fanIndex,
		slider:     gtk.NewScale(gtk.OrientationHorizontal, gtk.NewAdjustment(0, 0, 100, 5, 5, 0)),
		valueLabel: gtk.NewLabel("0%"),
		autoButton: gtk.NewToggleButton(),
	}

	box := gtk.NewBox(gtk.OrientationHorizontal, 5)
	control.slider.SetHExpand(true)
	control.slider.SetRoundDigits(0)
	control.slider.SetDigits(0)
	control.autoButton.SetLabel("Auto")

	box.Append(gtk.NewLabel(label))
	box.Append(control.slider)
	box.Append(control.valueLabel)
	box.Append(control.autoButton)

	setupFanControlCallbacks(control)
	return box, control
}

func setupFanControlCallbacks(control *FanControl) {
	control.slider.ConnectValueChanged(func() {
		if !control.autoButton.Active() {
			value := control.slider.Value()
			control.lastValue = value
			control.valueLabel.SetLabel(fmt.Sprintf("%.0f%%", value))
			go sendCommand(NBFCCommand{
				Command: "set-fan-speed",
				Fan:     control.fanIndex,
				Speed:   value,
			})
		}
	})

	control.autoButton.ConnectToggled(func() {
		if control.autoButton.Active() {
			control.lastValue = control.slider.Value()
			control.slider.SetSensitive(false)
			control.valueLabel.SetLabel("Auto")
			go sendCommand(NBFCCommand{Command: "set-fan-speed", Fan: control.fanIndex, Speed: "auto"})
		} else {
			control.slider.SetSensitive(true)
			control.slider.SetValue(control.lastValue)
			control.valueLabel.SetLabel(fmt.Sprintf("%.0f%%", control.lastValue))
			go sendCommand(NBFCCommand{Command: "set-fan-speed", Fan: control.fanIndex, Speed: control.lastValue})
		}
	})
}

func createStatusUI() (*gtk.Frame, *StatusLabels) {
	labels := &StatusLabels{
		cpuTemp:      gtk.NewLabel("--°C"),
		cpuFanSpeed:  gtk.NewLabel("Speed: --%"),
		cpuFanTarget: gtk.NewLabel("Target: --%"),
		cpuFanAuto:   gtk.NewLabel("Auto: --"),
		gpuFanSpeed:  gtk.NewLabel("Speed: --%"),
		gpuFanTarget: gtk.NewLabel("Target: --%"),
		gpuFanAuto:   gtk.NewLabel("Auto: --"),
	}

	frame := gtk.NewFrame("Thông tin hệ thống")
	box := gtk.NewBox(gtk.OrientationVertical, defaultSpacing)
	box.SetMarginTop(defaultSpacing)
	box.SetMarginBottom(defaultSpacing)
	box.SetMarginStart(defaultSpacing)
	box.SetMarginEnd(defaultSpacing)

	tempGrid := gtk.NewGrid()
	tempGrid.SetRowSpacing(defaultSpacing)
	tempGrid.SetColumnSpacing(defaultSpacing * 2)
	tempGrid.Attach(gtk.NewLabel("Nhiệt độ CPU:"), 0, 0, 1, 1)
	tempGrid.Attach(labels.cpuTemp, 1, 0, 1, 1)

	fanGrid := createFanStatusGrid(labels)

	box.Append(tempGrid)
	box.Append(fanGrid)
	frame.SetChild(box)

	return frame, labels
}

func createFanStatusGrid(labels *StatusLabels) *gtk.Grid {
	grid := gtk.NewGrid()
	grid.SetRowSpacing(defaultSpacing)
	grid.SetColumnSpacing(defaultSpacing * 2)
	grid.SetMarginTop(defaultSpacing)

	grid.Attach(gtk.NewLabel("CPU Fan Status:"), 0, 0, 2, 1)
	grid.Attach(labels.cpuFanSpeed, 0, 1, 1, 1)
	grid.Attach(labels.cpuFanTarget, 1, 1, 1, 1)
	grid.Attach(labels.cpuFanAuto, 0, 2, 2, 1)

	grid.Attach(gtk.NewLabel("GPU Fan Status:"), 2, 0, 2, 1)
	grid.Attach(labels.gpuFanSpeed, 2, 1, 1, 1)
	grid.Attach(labels.gpuFanTarget, 3, 1, 1, 1)
	grid.Attach(labels.gpuFanAuto, 2, 2, 2, 1)

	return grid
}

func updateStatus(labels *StatusLabels, controls []*FanControl) {
	output, err := exec.Command("bash", "-c",
		"echo '{\"command\":\"status\"}' | socat - UNIX-CONNECT:/var/run/nbfc_service.socket").CombinedOutput()
	if err != nil {
		return
	}

	lastBrace := bytes.LastIndex(output, []byte("}"))
	if lastBrace == -1 {
		return
	}

	var status NBFCStatus
	if err := json.Unmarshal(output[:lastBrace+1], &status); err != nil {
		return
	}

	glib.IdleAdd(func() bool {
		labels.cpuTemp.SetText(fmt.Sprintf("%.1f°C", status.Temperature))

		if len(status.Fans) > 0 {
			updateFanStatus(status.Fans[0], labels.cpuFanSpeed, labels.cpuFanTarget,
				labels.cpuFanAuto, controls[0])
		}
		if len(status.Fans) > 1 {
			updateFanStatus(status.Fans[1], labels.gpuFanSpeed, labels.gpuFanTarget,
				labels.gpuFanAuto, controls[1])
		}
		return false
	})
}

func updateFanStatus(fan Fan, speedLabel, targetLabel, autoLabel *gtk.Label, control *FanControl) {
	speedLabel.SetText(fmt.Sprintf("Speed: %.1f%%", fan.CurrentSpeed))
	targetLabel.SetText(fmt.Sprintf("Target: %.1f%%", fan.TargetSpeed))
	autoLabel.SetText(fmt.Sprintf("Auto: %v", fan.AutoMode))

	if fan.AutoMode != control.autoButton.Active() {
		control.autoButton.SetActive(fan.AutoMode)
	}
}

func activate(app *gtk.Application) {
	// Load cấu hình GTK
	gtkConfig := loadGtkConfig()

	// Áp dụng cấu hình
	settings := gtk.SettingsGetDefault()
	if theme, ok := gtkConfig["gtk-theme-name"]; ok {
		settings.SetObjectProperty("gtk-theme-name", theme)
	}
	if iconTheme, ok := gtkConfig["icon-theme"]; ok {
		settings.SetObjectProperty("gtk-icon-theme-name", iconTheme)
	}
	if darkTheme, ok := gtkConfig["gtk-application-prefer-dark-theme"]; ok {
		settings.SetObjectProperty("gtk-application-prefer-dark-theme", darkTheme == "1")
	}

	cssProvider := gtk.NewCSSProvider()
	css := `
	.background {
		background-color: @theme_bg_color;
	}
	window, frame, box {
		background-color: @theme_bg_color;
		color: @theme_fg_color;
		border-radius: 12px;
	}
	button {
		background-color: @theme_button_bg_color;
		color: @theme_button_fg_color;
		padding: 8px;
		border-radius: 6px;
	}
	frame {
		padding: 12px;
		margin: 8px;
	}
	scale {
		margin: 8px;
		min-height: 24px;
	}
	label {
		margin: 4px;
	}
	`

	cssProvider.LoadFromString(css)
	display := gdk.DisplayGetDefault()
	gtk.StyleContextAddProviderForDisplay(
		display,
		cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	window := gtk.NewApplicationWindow(app)
	window.AddCSSClass("background")
	window.SetTitle("NBFC GUI Controller")
	window.SetDefaultSize(windowWidth, windowHeight)

	mainBox := gtk.NewBox(gtk.OrientationVertical, defaultSpacing)
	mainBox.SetMarginTop(defaultMargin)
	mainBox.SetMarginBottom(defaultMargin)
	mainBox.SetMarginStart(defaultMargin)
	mainBox.SetMarginEnd(defaultMargin)

	controlsFrame := gtk.NewFrame("Điều khiển quạt")
	controlsBox := gtk.NewBox(gtk.OrientationVertical, defaultSpacing)
	controlsBox.SetMarginTop(defaultSpacing)
	controlsBox.SetMarginBottom(defaultSpacing)
	controlsBox.SetMarginStart(defaultSpacing)
	controlsBox.SetMarginEnd(defaultSpacing)

	cpuFanBox, cpuControl := createFanControl("CPU Fan:", 0)
	gpuFanBox, gpuControl := createFanControl("GPU Fan:", 1)
	controls := []*FanControl{cpuControl, gpuControl}

	controlsBox.Append(cpuFanBox)
	controlsBox.Append(gpuFanBox)
	controlsFrame.SetChild(controlsBox)

	statusFrame, labels := createStatusUI()

	refreshButton := gtk.NewButtonWithLabel("Cập nhật trạng thái")
	refreshButton.ConnectClicked(
		func() {
			go updateStatus(labels, controls)
		})

	mainBox.Append(controlsFrame)
	mainBox.Append(statusFrame)
	mainBox.Append(refreshButton)

	window.SetChild(mainBox)
	window.SetVisible(true)
}

func main() {
	app := gtk.NewApplication("com.nbfc.nbfc-gui", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		activate(app)
	})
	if code := app.Run(nil); code > 0 {
		panic(code)
	}
}
