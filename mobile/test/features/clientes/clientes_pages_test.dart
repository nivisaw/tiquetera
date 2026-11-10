import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';

import 'fixtures.dart';

ClientePagina _page(List<Cliente> items, {int total = 1, int page = 1}) =>
    ClientePagina(items: items, page: page, total: total);

void main() {
  setUpAll(() => registerFallbackValue(DateTime(2026)));

  testWidgets('TestClientesPage_List_ShowsItemsAndLoadsMore', (tester) async {
    final repo = MockClientesRepo();
    when(repo.list).thenAnswer((_) async => _page([cliente], total: 2));
    when(() => repo.list(page: 2)).thenAnswer(
      (_) async => _page(
        [cliente.copyWith(id: 'c2', nombre: 'Luis')],
        total: 2,
        page: 2,
      ),
    );
    await tester.pumpWidget(harness('/clientes', repo));
    await tester.pumpAndSettle();

    expect(find.text('Ana Pérez'), findsOneWidget);
    expect(find.text('300 123 4567'), findsOneWidget);
    await tester.tap(find.byKey(const Key('clientes_more')));
    await tester.pumpAndSettle();
    expect(find.text('Luis'), findsOneWidget);
    expect(find.byKey(const Key('clientes_more')), findsNothing);
  });

  testWidgets('TestClientesPage_Search_DebouncesAndShowsNoResults', (
    tester,
  ) async {
    final repo = MockClientesRepo();
    when(repo.list).thenAnswer((_) async => _page([cliente]));
    when(() => repo.list(query: 'zz'))
        .thenAnswer((_) async => _page([], total: 0));
    await tester.pumpWidget(harness('/clientes', repo));
    await tester.pumpAndSettle();

    await tester.enterText(find.byKey(const Key('clientes_search')), 'z');
    await tester.enterText(find.byKey(const Key('clientes_search')), 'zz');
    await tester.pump(const Duration(milliseconds: 450));
    await tester.pumpAndSettle();

    verify(() => repo.list(query: 'zz')).called(1);
    verifyNever(() => repo.list(query: 'z'));
    expect(
      find.text('No hay clientes que coincidan con la búsqueda.'),
      findsOneWidget,
    );
  });

  testWidgets('TestClientesPage_Error_RetryReloads', (tester) async {
    final repo = MockClientesRepo();
    var calls = 0;
    when(repo.list).thenAnswer((_) async {
      calls++;
      if (calls == 1) throw const NetworkFailure();
      return _page([], total: 0);
    });
    await tester.pumpWidget(harness('/clientes', repo));
    await tester.pumpAndSettle();
    expect(
      find.text('Sin conexión. Verifica tu red e inténtalo de nuevo.'),
      findsOneWidget,
    );
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(find.text('No hay clientes registrados.'), findsOneWidget);
  });

  testWidgets('TestDetalle_Regenerar_CancelDoesNothingConfirmOpensQr', (
    tester,
  ) async {
    final repo = MockClientesRepo();
    when(() => repo.get('c1')).thenAnswer((_) async => cliente);
    when(() => repo.regenerarQr('c1')).thenAnswer((_) async => 'TQ1.new');
    await tester.pumpWidget(harness('/clientes/c1', repo));
    await tester.pumpAndSettle();
    expect(find.text('20 almuerzos disponibles'), findsOneWidget);

    await tester.tap(find.byKey(const Key('detalle_regenerar')));
    await tester.pumpAndSettle();
    expect(find.text('¿Regenerar el código QR?'), findsOneWidget);
    await tester.tap(find.text('Cancelar'));
    await tester.pumpAndSettle();
    verifyNever(() => repo.regenerarQr(any()));

    await tester.tap(find.byKey(const Key('detalle_regenerar')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('regenerar_confirm')));
    await tester.pumpAndSettle();
    verify(() => repo.regenerarQr('c1')).called(1);
    final qr = tester.widget<QrPage>(find.byType(QrPage));
    expect(qr.args.payload, 'TQ1.new');
  });

  testWidgets('TestDetalle_Recargar_SendsAndRefreshes', (tester) async {
    final repo = MockClientesRepo();
    var saldo = 20;
    when(() => repo.get('c1')).thenAnswer(
      (_) async => cliente.copyWith(
        tiquetera: cliente.tiquetera!.copyWith(saldo: saldo),
      ),
    );
    when(
      () => repo.recargar(
        'c1',
        almuerzos: any(named: 'almuerzos'),
        fechaVencimiento: any(named: 'fechaVencimiento'),
      ),
    ).thenAnswer((_) async {
      saldo = 40;
      return cliente.tiquetera!.copyWith(saldo: 40);
    });
    await tester.pumpWidget(harness('/clientes/c1', repo));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('detalle_recargar')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('recarga_save')));
    await tester.pumpAndSettle();

    verify(
      () => repo.recargar(
        'c1',
        almuerzos: 20,
        fechaVencimiento: DateTime(2026, 12, 10),
      ),
    ).called(1);
    expect(find.text('40 almuerzos disponibles'), findsOneWidget);
    expect(find.text('Recarga registrada.'), findsOneWidget);
  });

  testWidgets('TestDetalle_RecargaInactivo_ShowsError', (tester) async {
    final repo = MockClientesRepo();
    when(() => repo.get('c1')).thenAnswer((_) async => cliente);
    when(
      () => repo.recargar(
        'c1',
        almuerzos: any(named: 'almuerzos'),
        fechaVencimiento: any(named: 'fechaVencimiento'),
      ),
    ).thenThrow(const ApiFailure(status: 422, code: 'CLIENTE_INACTIVO'));
    await tester.pumpWidget(harness('/clientes/c1', repo));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('detalle_recargar')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('recarga_save')));
    await tester.pumpAndSettle();
    expect(find.text('El cliente está inactivo.'), findsOneWidget);
  });
}
