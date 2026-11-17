import 'dart:io';
import 'dart:typed_data';

import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';

typedef ShareFn = Future<ShareResult> Function(ShareParams params);
typedef LaunchFn = Future<bool> Function(Uri uri);
typedef TempDirFn = Future<Directory> Function();

/// Comparte el PNG del QR y abre chats de WhatsApp. El archivo temporal se
/// borra siempre al terminar, aunque el usuario cancele o falle el envío.
class QrShareService {
  const new({ShareFn? share, LaunchFn? launch, TempDirFn? tempDir})
    : _share = share ?? _defaultShare,
      _launch = launch ?? _defaultLaunch,
      _tempDir = tempDir ?? getTemporaryDirectory;

  final ShareFn _share;
  final LaunchFn _launch;
  final TempDirFn _tempDir;

  static Future<ShareResult> _defaultShare(ShareParams p) =>
      SharePlus.instance.share(p);

  static Future<bool> _defaultLaunch(Uri uri) =>
      launchUrl(uri, mode: LaunchMode.externalApplication);

  Future<ShareResult> sharePng(Uint8List png, {required String text}) async {
    final dir = await _tempDir();
    final file = File(
      '${dir.path}/qr_${DateTime.now().microsecondsSinceEpoch}.png',
    );
    await file.writeAsBytes(png, flush: true);
    try {
      return await _share(
        ShareParams(
          text: text,
          files: [XFile(file.path, mimeType: 'image/png')],
        ),
      );
    } finally {
      if (file.existsSync()) await file.delete();
    }
  }

  Future<bool> openChat(Uri uri) => _launch(uri);
}
