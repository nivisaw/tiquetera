import 'package:flutter/services.dart';

/// Mantiene la pantalla encendida (canal nativo en MainActivity/AppDelegate).
abstract final class ScreenAwake {
  static const MethodChannel channel = MethodChannel('tiquetera/screen');

  static Future<void> set({required bool on}) async {
    try {
      await channel.invokeMethod<void>('keepOn', on);
    } on PlatformException {
      return;
    } on MissingPluginException {
      return;
    }
  }
}
