import 'package:permission_handler/permission_handler.dart';

enum CameraAccess { granted, denied, permanentlyDenied }

/// Envoltorio de `permission_handler` para poder simularlo en tests.
class CameraPermission {
  const new();

  Future<CameraAccess> status() async =>
      fromStatus(await Permission.camera.status);

  Future<CameraAccess> request() async =>
      fromStatus(await Permission.camera.request());

  Future<bool> openSettings() => openAppSettings();

  static CameraAccess fromStatus(PermissionStatus s) => switch (s) {
    PermissionStatus.granted ||
    PermissionStatus.limited => CameraAccess.granted,
    PermissionStatus.permanentlyDenied ||
    PermissionStatus.restricted => CameraAccess.permanentlyDenied,
    _ => CameraAccess.denied,
  };
}
