import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';

import 'fixtures.dart';

Consumo _c(String id, int saldo) => Consumo(
  id: id,
  fechaHora: DateTime.utc(2026, 11, 10, 17, 5),
  saldoResultante: saldo,
);

ConsumoPagina _p(List<Consumo> items, {int total = 1, int page = 1}) =>
    ConsumoPagina(items: items, page: page, total: total);

MockClientesRepo _repo() {
  final repo = MockClientesRepo();
  when(() => repo.get('c1')).thenAnswer((_) async => cliente);
  return repo;
}

Future<void> _open(WidgetTester tester, MockClientesRepo repo) async {
  tester.view.physicalSize = const Size(800, 3000);
  addTearDown(tester.view.reset);
  await tester.pumpWidget(harness('/clientes/c1', repo));
  await tester.pumpAndSettle();
}

void main() {
  test('TestConsumo_FromJson_ParsesUtc', () {
    final c = Consumo.fromJson({
      'id': 'k1',
      'tiquetera_id': 't1',
      'cliente_id': 'c1',
      'cajero_id': 'u1',
      'fecha_hora': '2026-11-10T12:05:00-05:00',
      'saldo_resultante': 19,
    });
    expect(c.fechaHora, DateTime.utc(2026, 11, 10, 17, 5));
    expect(c.saldoResultante, 19);
  });

  test('TestClienteRules_DisplayDateTimeBogota_ConvertsOffset', () {
    expect(
      ClienteRules.displayDateTimeBogota(DateTime.utc(2026, 11, 11, 3, 7)),
      '10/11/2026 22:07',
    );
  });

  testWidgets('TestHistorial_Items_ShowsAndLoadsMore', (tester) async {
    final repo = _repo();
    when(() => repo.listConsumos('c1'))
        .thenAnswer((_) async => _p([_c('k1', 19)], total: 2));
    when(() => repo.listConsumos('c1', page: 2))
        .thenAnswer((_) async => _p([_c('k2', 1)], total: 2, page: 2));
    await _open(tester, repo);

    expect(find.text('10/11/2026 12:05'), findsOneWidget);
    expect(find.text('Quedaron 19 almuerzos'), findsOneWidget);
    await tester.tap(find.byKey(const Key('historial_more')));
    await tester.pumpAndSettle();
    expect(find.text('Quedó 1 almuerzo'), findsOneWidget);
    expect(find.byKey(const Key('historial_more')), findsNothing);
  });

  testWidgets('TestHistorial_Empty_ShowsMessage', (tester) async {
    final repo = _repo();
    when(() => repo.listConsumos('c1'))
        .thenAnswer((_) async => _p([], total: 0));
    await _open(tester, repo);
    expect(find.text('Aún no hay consumos registrados.'), findsOneWidget);
  });

  testWidgets('TestHistorial_Error_RetryReloads', (tester) async {
    final repo = _repo();
    var calls = 0;
    when(() => repo.listConsumos('c1')).thenAnswer((_) async {
      if (calls++ == 0) throw const NetworkFailure();
      return _p([_c('k1', 5)]);
    });
    await _open(tester, repo);
    expect(find.text('No se pudo cargar el historial.'), findsOneWidget);

    await tester.tap(find.byKey(const Key('historial_retry')));
    await tester.pumpAndSettle();
    expect(find.text('Quedaron 5 almuerzos'), findsOneWidget);
  });

  testWidgets('TestHistorial_LoadMoreFails_KeepsItems', (tester) async {
    final repo = _repo();
    when(() => repo.listConsumos('c1'))
        .thenAnswer((_) async => _p([_c('k1', 19)], total: 2));
    when(() => repo.listConsumos('c1', page: 2))
        .thenThrow(const NetworkFailure());
    await _open(tester, repo);
    await tester.tap(find.byKey(const Key('historial_more')));
    await tester.pumpAndSettle();
    expect(find.text('Quedaron 19 almuerzos'), findsOneWidget);
    expect(find.byKey(const Key('historial_more')), findsOneWidget);
  });
}
