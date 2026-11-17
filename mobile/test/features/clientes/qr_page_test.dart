import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:share_plus/share_plus.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/data/qr_share_service.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';

import 'fixtures.dart';

final _args = QrArgs(
  clienteId: 'c1',
  nombre: 'Ana Pérez',
  saldo: 20,
  fechaVencimiento: DateTime(2026, 12, 10),
  telefono: '+573001234567',
  payload: 'TQ1.AbCdEfGhIjKlMnOpQrStUvWxYz0123456789-_AbCd',
);

void main() {
  shareTests();

  testWidgets('TestQrPage_Shows_NameSaldoVencimientoWithoutToken', (
    tester,
  ) async {
    await tester.pumpWidget(qrHarness(_args));
    expect(find.text('Ana Pérez'), findsOneWidget);
    expect(find.text('20 almuerzos disponibles'), findsOneWidget);
    expect(find.text('Vence el 10/12/2026'), findsOneWidget);
    expect(find.textContaining('TQ1.'), findsNothing);
    expect(find.bySemanticsLabel('Código QR de Ana Pérez'), findsOneWidget);
  });

  testWidgets('TestQrPage_Light_MatchesGolden', (tester) async {
    tester.view
      ..physicalSize = const Size(1080, 2340)
      ..devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(qrHarness(_args));
    await expectLater(
      find.byType(QrPage),
      matchesGoldenFile('goldens/qr_page_light.png'),
    );
  });

  testWidgets('TestQrPage_Dark_KeepsWhiteQrAndMatchesGolden', (tester) async {
    tester.view
      ..physicalSize = const Size(1080, 2340)
      ..devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(qrHarness(_args, theme: AppTheme.dark()));
    final card = tester.widget<Container>(find.byKey(const Key('qr_card')));
    expect(card.color, Colors.white);
    await expectLater(
      find.byType(QrPage),
      matchesGoldenFile('goldens/qr_page_dark.png'),
    );
  });
}

class _FakeShare extends QrShareService {
  new({this.chatOk = true, this.shareThrows = false});

  final bool chatOk;
  final bool shareThrows;
  String? sharedText;
  Uri? openedUri;

  @override
  Future<ShareResult> sharePng(Uint8List png, {required String text}) async {
    if (shareThrows) throw Exception('boom');
    sharedText = text;
    return const ShareResult('ok', ShareResultStatus.success);
  }

  @override
  Future<bool> openChat(Uri uri) async {
    openedUri = uri;
    return chatOk;
  }
}

Future<void> _tapKey(WidgetTester tester, String key) async {
  final finder = find.byKey(Key(key));
  await tester.ensureVisible(finder);
  await tester.tap(finder);
}

void shareTests() {
  testWidgets('TestQrPage_Share_SendsPngWithMessage', (tester) async {
    final fake = _FakeShare();
    await tester.pumpWidget(qrHarness(_args, share: fake));
    await tester.runAsync(() async {
      await _tapKey(tester, 'qr_share');
      await Future<void>.delayed(const Duration(milliseconds: 200));
    });
    await tester.pump();
    expect(fake.sharedText, contains('Ana Pérez'));
    expect(fake.sharedText, contains('20 almuerzos'));
    expect(fake.sharedText, contains('10/12/2026'));
    expect(fake.sharedText, isNot(contains('TQ1.')));
  });

  testWidgets('TestQrPage_ShareFails_ShowsError', (tester) async {
    await tester.pumpWidget(
      qrHarness(_args, share: _FakeShare(shareThrows: true)),
    );
    await tester.runAsync(() async {
      await _tapKey(tester, 'qr_share');
      await Future<void>.delayed(const Duration(milliseconds: 200));
    });
    await tester.pump();
    expect(
      find.text('No se pudo compartir el código QR. Intenta de nuevo.'),
      findsOneWidget,
    );
  });

  testWidgets('TestQrPage_OpenChat_OpensWaMeWithClientPhone', (tester) async {
    final fake = _FakeShare();
    await tester.pumpWidget(qrHarness(_args, share: fake));
    await _tapKey(tester, 'qr_open_chat');
    await tester.pump();
    expect(fake.openedUri!.host, 'wa.me');
    expect(fake.openedUri!.path, '/573001234567');
    expect(fake.openedUri!.queryParameters['text'], contains('Ana Pérez'));
  });

  testWidgets('TestQrPage_OpenChatFails_ShowsError', (tester) async {
    await tester.pumpWidget(qrHarness(_args, share: _FakeShare(chatOk: false)));
    await _tapKey(tester, 'qr_open_chat');
    await tester.pump();
    expect(find.text('No se pudo abrir WhatsApp.'), findsOneWidget);
  });
}
