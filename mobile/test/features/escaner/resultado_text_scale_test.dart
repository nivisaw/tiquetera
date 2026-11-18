import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';
import 'package:tiquetera/features/escaner/presentation/resultado_view.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

Widget _app(Widget child) => MaterialApp(
  theme: AppTheme.light(),
  locale: const Locale('es', 'CO'),
  supportedLocales: AppLocalizations.supportedLocales,
  localizationsDelegates: AppLocalizations.localizationsDelegates,
  home: Scaffold(body: child),
);

void main() {
  testWidgets('TestResultadoView_TextScale200_NoOverflow', (tester) async {
    tester.view
      ..physicalSize = const Size(360, 640)
      ..devicePixelRatio = 1;
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(() {
      tester.view.reset();
      tester.platformDispatcher.clearTextScaleFactorTestValue();
    });
    await tester.pumpWidget(
      _app(
        ResultadoView(
          onDone: () {},
          consumo: ConsumoOk(
            nombre: 'Ana María Pérez Rodríguez',
            saldoRestante: 500,
            fechaVencimiento: DateTime(2026, 12, 31),
          ),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 600));
    expect(tester.takeException(), isNull);
    await tester.pump(const Duration(seconds: 3));
  });
}
