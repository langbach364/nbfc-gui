{pkgs, lib, ... }:

let
  package = pkgs.stdenv.mkDerivation {
    pname = "nbfc-gui";
    version = "1.0.0";
    src = ../.;

    nativeBuildInputs = [ pkgs.makeWrapper ];

    installPhase = ''
      mkdir -p $out/bin
      mkdir -p $out/share/applications
      mkdir -p $out/share/icons
      
      cp bin/nbfc-gui $out/bin/
      cp desktop/fan.desktop $out/share/applications/nbfc-gui.desktop
      cp png/fan.png $out/share/icons/fan.png
      
      chmod +x $out/bin/nbfc-gui

      wrapProgram $out/bin/nbfc-gui \
        --prefix LD_LIBRARY_PATH : "${pkgs.gtk4}/lib"
    '';

    meta = with lib; {
      description = "GUI Controller for Notebook Fan Control";
      platforms = platforms.linux;
    };
  };
in
{
  imports = [ 
    ../module/GTK.nix
  ];
  
  home.packages = with pkgs; [
    socat
    package
  ];
}
