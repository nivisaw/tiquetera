import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';

import 'fixtures.dart';

final _args = QrArgs(
  clienteId: 'c1',
  nombre: 'Ana Pérez',
  saldo: 20,
  fechaVencimiento: DateTime(2026, 12, 10),
  payload: 'TQ1.AbCdEfGhIjKlMnOpQrStUvWxYz0123456789-_AbCd',
);

void main() {
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
