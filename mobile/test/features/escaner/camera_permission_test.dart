import 'package:flutter_test/flutter_test.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:tiquetera/features/escaner/data/camera_permission.dart';

void main() {
  test('TestCameraPermission_FromStatus_MapsAllStatuses', () {
    const expected = {
      PermissionStatus.granted: CameraAccess.granted,
      PermissionStatus.limited: CameraAccess.granted,
      PermissionStatus.denied: CameraAccess.denied,
      PermissionStatus.provisional: CameraAccess.denied,
      PermissionStatus.permanentlyDenied: CameraAccess.permanentlyDenied,
      PermissionStatus.restricted: CameraAccess.permanentlyDenied,
    };
    for (final e in expected.entries) {
      expect(CameraPermission.fromStatus(e.key), e.value, reason: '${e.key}');
    }
  });
}
