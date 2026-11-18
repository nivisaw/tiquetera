import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';

import 'fixtures.dart';

/// Teléfono pequeño (360x640 dp) con texto al 200%: sin desbordes.
void _smallPhoneLargeText(WidgetTester tester) {
  tester.view
    ..physicalSize = const Size(360, 640)
    ..devicePixelRatio = 1;
  tester.platformDispatcher.textScaleFactorTestValue = 2;
  addTearDown(() {
    tester.view.reset();
    tester.platformDispatcher.clearTextScaleFactorTestValue();
  });
}

MockClientesRepo _repo() {
  final repo = MockClientesRepo();
  when(
    repo.list,
  ).thenAnswer((_) async => ClientePagina(items: [cliente], page: 1, total: 2));
  when(() => repo.get('c1')).thenAnswer((_) async => cliente);
  when(() => repo.listConsumos('c1')).thenAnswer(
    (_) async => ConsumoPagina(
      items: [
        Consumo(
          id: 'k1',
          fechaHora: DateTime.utc(2026, 11, 10, 17),
          saldoResultante: 19,
        ),
      ],
      page: 1,
      total: 2,
    ),
  );
  return repo;
}

void main() {
  for (final route in ['/clientes', '/clientes/nuevo', '/clientes/c1']) {
    testWidgets('TestTextScale200_${route}_NoOverflow', (tester) async {
      _smallPhoneLargeText(tester);
      await tester.pumpWidget(harness(route, _repo()));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('TestTextScale200_QrPage_NoOverflow', (tester) async {
    _smallPhoneLargeText(tester);
    await tester.pumpWidget(
      qrHarness(
        QrArgs(
          clienteId: 'c1',
          nombre: 'Ana María Pérez Rodríguez',
          saldo: 500,
          fechaVencimiento: DateTime(2026, 12, 10),
          telefono: '+573001234567',
          payload: 'TQ1.x',
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
