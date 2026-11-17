import 'package:flutter_riverpod/misc.dart';
import 'package:tiquetera/features/escaner/data/camera_permission.dart';
import 'package:tiquetera/features/escaner/presentation/escaner_controller.dart';

class _DeniedCamera implements CameraPermission {
  const new();

  @override
  Future<CameraAccess> status() async => CameraAccess.denied;

  @override
  Future<CameraAccess> request() async => CameraAccess.denied;

  @override
  Future<bool> openSettings() async => false;
}

/// Evita el canal nativo de permisos en tests que navegan al escáner.
final Override deniedCameraOverride = cameraPermissionProvider
    .overrideWithValue(const _DeniedCamera());
