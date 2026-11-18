import Flutter
import UIKit

@main
@objc class AppDelegate: FlutterAppDelegate, FlutterImplicitEngineDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }

  func didInitializeImplicitFlutterEngine(_ engineBridge: FlutterImplicitEngineBridge) {
    GeneratedPluginRegistrant.register(with: engineBridge.pluginRegistry)
    guard let registrar = engineBridge.pluginRegistry.registrar(forPlugin: "ScreenAwake") else { return }
    FlutterMethodChannel(name: "tiquetera/screen", binaryMessenger: registrar.messenger())
      .setMethodCallHandler { call, result in
        guard call.method == "keepOn" else { return result(FlutterMethodNotImplemented) }
        UIApplication.shared.isIdleTimerDisabled = (call.arguments as? Bool) ?? false
        result(nil)
      }
  }
}
