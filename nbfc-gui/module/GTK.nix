{ pkgs, lib, ... }:

{
  home.packages = with pkgs; [
    pkg-config
    glib.dev
    gtk4.dev
    gobject-introspection.dev
    cairo.dev
    gcc
    libgtkflow4
    pango.dev
    gdk-pixbuf.dev
    harfbuzz.dev
    vulkan-loader
    graphene.dev
  ];

  home.activation = {
    linkPkgConfig = lib.hm.dag.entryAfter ["writeBoundary"] ''
      mkdir -p $HOME/.local/lib/pkgconfig
      ln -sf ${pkgs.glib.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.gtk4.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/ 
      ln -sf ${pkgs.gobject-introspection.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.cairo.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.pango.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.gdk-pixbuf.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.harfbuzz.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.vulkan-loader.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
      ln -sf ${pkgs.graphene.dev}/lib/pkgconfig/* $HOME/.local/lib/pkgconfig/
    '';
  };
}
