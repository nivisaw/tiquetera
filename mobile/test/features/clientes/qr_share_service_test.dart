import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:share_plus/share_plus.dart';
import 'package:tiquetera/features/clientes/data/qr_share_service.dart';

void main() {
  late Directory tmp;

  setUp(() => tmp = Directory.systemTemp.createTempSync('qr_share_test'));
  tearDown(() => tmp.deleteSync(recursive: true));

  test('TestQrShareService_SharePng_SharesFileAndDeletesIt', () async {
    late ShareParams sent;
    var existedWhileSharing = false;
    final service = QrShareService(
      tempDir: () async => tmp,
      share: (p) async {
        sent = p;
        existedWhileSharing = File(p.files!.single.path).existsSync();
        return const ShareResult('ok', ShareResultStatus.success);
      },
    );
    final png = Uint8List.fromList([1, 2, 3]);

    final r = await service.sharePng(png, text: 'hola');

    expect(r.status, ShareResultStatus.success);
    expect(sent.text, 'hola');
    expect(sent.files!.single.mimeType, 'image/png');
    expect(existedWhileSharing, isTrue);
    expect(tmp.listSync(), isEmpty);
  });

  test('TestQrShareService_ShareFails_StillDeletesFile', () async {
    final service = QrShareService(
      tempDir: () async => tmp,
      share: (_) async => throw Exception('boom'),
    );

    await expectLater(
      service.sharePng(Uint8List(1), text: 'x'),
      throwsException,
    );
    expect(tmp.listSync(), isEmpty);
  });

  test('TestQrShareService_OpenChat_DelegatesToLauncher', () async {
    Uri? opened;
    final service = QrShareService(
      launch: (u) async {
        opened = u;
        return true;
      },
    );
    final uri = Uri.parse('https://wa.me/573001234567');

    expect(await service.openChat(uri), isTrue);
    expect(opened, uri);
  });
}
