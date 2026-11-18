import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:mocktail/mocktail.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/data/qr_share_service.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/reportes/data/reportes_repository.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';
import 'package:tiquetera/features/reportes/presentation/consumos_reporte_page.dart';
import 'package:tiquetera/features/reportes/presentation/reporte_clientes_page.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_page.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_providers.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class MockReportesRepo extends Mock implements ReportesRepository;

/// 2026-11-07 10:00 en Bogotá.
final DateTime _now = DateTime.utc(2026, 11, 7, 15);

final Dashboard dashboardFixture = Dashboard(
  consumosHoy: 12,
  consumos7d: [
    for (var i = 1; i <= 7; i++)
      DiaTotal(fecha: DateTime(2026, 11, i), total: i),
  ],
  clientesVigentes: 34,
  saldoPendiente: 410,
  vendidosMes: 95,
  porVencer7d: 5,
  rechazosHoy: 2,
);

ClienteReporte reporteItem(String id, {int saldo = 2}) => ClienteReporte(
  clienteId: id,
  nombre: 'Ana Pérez',
  telefono: '+573001234567',
  fecha: DateTime(2026, 11, 10),
  saldo: saldo,
);

ReportePagina pagina(
  List<ClienteReporte> items, {
  int total = 1,
  int page = 1,
}) => ReportePagina(items: items, page: page, total: total);

Widget reportesHarness(
  String initial,
  MockReportesRepo repo, {
  List<Uri>? launched,
}) {
  final router = GoRouter(
    initialLocation: initial,
    routes: [
      GoRoute(
        path: '/reportes',
        builder: (_, _) => const ReportesPage(),
        routes: [
          GoRoute(
            path: 'consumos',
            builder: (_, _) => const ConsumosReportePage(),
          ),
          GoRoute(
            path: ':tipo',
            builder: (_, s) => ReporteClientesPage(
              tipo: ReporteTipo.fromPath(s.pathParameters['tipo']!)!,
            ),
          ),
        ],
      ),
      GoRoute(
        path: '/clientes/:id',
        builder: (_, s) => Text('detalle ${s.pathParameters['id']}'),
      ),
    ],
  );
  return ProviderScope(
    overrides: [
      reportesRepositoryProvider.overrideWithValue(repo),
      clockProvider.overrideWithValue(() => _now),
      qrShareServiceProvider.overrideWithValue(
        QrShareService(
          launch: (uri) async {
            launched?.add(uri);
            return launched != null;
          },
        ),
      ),
    ],
    child: MaterialApp.router(
      theme: AppTheme.light(),
      locale: const Locale('es', 'CO'),
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      routerConfig: router,
    ),
  );
}

void _tall(WidgetTester tester) {
  tester.view.physicalSize = const Size(800, 2400);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
}

void main() {
  setUpAll(() {
    registerFallbackValue(ReporteTipo.sinRecarga);
    registerFallbackValue(RangoFechas(DateTime(2026), DateTime(2026)));
  });

  testWidgets('TestReportesPage_Dashboard_ShowsFiguresAndLinks', (
    tester,
  ) async {
    _tall(tester);
    final repo = MockReportesRepo();
    when(repo.dashboard).thenAnswer((_) async => dashboardFixture);
    await tester.pumpWidget(reportesHarness('/reportes', repo));
    await tester.pumpAndSettle();

    expect(find.text('410'), findsOneWidget);
    expect(find.text('Almuerzos prepagados por servir'), findsOneWidget);
    final semantics = tester.ensureSemantics();
    expect(find.bySemanticsLabel('07/11/2026: 7 almuerzos'), findsOneWidget);
    expect(find.bySemanticsLabel('Almuerzos servidos hoy: 12'), findsOneWidget);
    semantics.dispose();
    expect(find.text('Sin recarga'), findsOneWidget);
    expect(find.text('Consumos por día'), findsOneWidget);
  });

  testWidgets('TestReportesPage_Error_RetryReloads', (tester) async {
    _tall(tester);
    final repo = MockReportesRepo();
    var calls = 0;
    when(repo.dashboard).thenAnswer((_) async {
      if (calls++ == 0) throw const NetworkFailure();
      return dashboardFixture;
    });
    await tester.pumpWidget(reportesHarness('/reportes', repo));
    await tester.pumpAndSettle();
    expect(find.text('Reintentar'), findsOneWidget);

    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(find.text('410'), findsOneWidget);
  });

  testWidgets('TestReporteClientes_FilterChangesQueryAndOpensDetail', (
    tester,
  ) async {
    final repo = MockReportesRepo();
    when(() => repo.list(ReporteTipo.sinRecarga, 30))
        .thenAnswer((_) async => pagina([reporteItem('c1')]));
    when(() => repo.list(ReporteTipo.sinRecarga, 60))
        .thenAnswer((_) async => pagina([], total: 0));
    await tester.pumpWidget(reportesHarness('/reportes/sin-recarga', repo));
    await tester.pumpAndSettle();

    expect(find.textContaining('Última recarga: 10/11/2026'), findsOneWidget);
    expect(find.text('1 cliente'), findsOneWidget);

    await tester.tap(find.byKey(const Key('filtro_60')));
    await tester.pumpAndSettle();
    expect(find.text('No hay clientes en este reporte.'), findsOneWidget);

    await tester.tap(find.byKey(const Key('filtro_30')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Ana Pérez'));
    await tester.pumpAndSettle();
    expect(find.text('detalle c1'), findsOneWidget);
  });

  testWidgets('TestReporteClientes_Chat_OpensWhatsAppWithMessage', (
    tester,
  ) async {
    final repo = MockReportesRepo();
    final launched = <Uri>[];
    when(() => repo.list(ReporteTipo.porVencer, 7))
        .thenAnswer((_) async => pagina([reporteItem('c1', saldo: 4)]));
    await tester.pumpWidget(
      reportesHarness('/reportes/por-vencer', repo, launched: launched),
    );
    await tester.pumpAndSettle();
    expect(find.textContaining('Vence: 10/11/2026'), findsOneWidget);

    await tester.tap(find.byKey(const Key('chat_c1')));
    await tester.pumpAndSettle();

    final uri = launched.single;
    expect(uri.host, 'wa.me');
    expect(uri.path, '/573001234567');
    expect(uri.queryParameters['text'], contains('Hola Ana'));
    expect(uri.queryParameters['text'], contains('4 almuerzos'));
  });

  testWidgets('TestReporteClientes_ChatFails_ShowsError', (tester) async {
    final repo = MockReportesRepo();
    when(() => repo.list(ReporteTipo.saldoBajo, 3))
        .thenAnswer((_) async => pagina([reporteItem('c1')]));
    await tester.pumpWidget(reportesHarness('/reportes/saldo-bajo', repo));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('chat_c1')));
    await tester.pumpAndSettle();
    expect(find.text('No se pudo abrir WhatsApp.'), findsOneWidget);
  });

  testWidgets('TestReporteClientes_LoadMoreAndErrorRetry', (tester) async {
    final repo = MockReportesRepo();
    var calls = 0;
    when(() => repo.list(ReporteTipo.saldoBajo, 3)).thenAnswer((_) async {
      if (calls++ == 0) throw const NetworkFailure();
      return pagina([reporteItem('c1')], total: 2);
    });
    when(
      () => repo.list(ReporteTipo.saldoBajo, 3, page: 2),
    ).thenAnswer((_) async => pagina([reporteItem('c2')], total: 2, page: 2));
    await tester.pumpWidget(reportesHarness('/reportes/saldo-bajo', repo));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('reporte_more')));
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('reporte_c2')), findsOneWidget);
    expect(find.byKey(const Key('reporte_more')), findsNothing);
  });

  testWidgets('TestConsumosReporte_DefaultWeek_ShowsTotalsAndCashiers', (
    tester,
  ) async {
    _tall(tester);
    final repo = MockReportesRepo();
    final semana = RangoFechas(DateTime(2026, 11), DateTime(2026, 11, 7));
    when(() => repo.consumos(semana)).thenAnswer(
      (_) async => ConsumosReporte(
        total: 28,
        dias: dashboardFixture.consumos7d,
        cajeros: const [CajeroTotal(nombre: 'Caja 1', total: 28)],
      ),
    );
    await tester.pumpWidget(reportesHarness('/reportes/consumos', repo));
    await tester.pumpAndSettle();

    expect(find.text('Del 01/11/2026 al 07/11/2026'), findsOneWidget);
    expect(find.text('28 almuerzos servidos'), findsOneWidget);
    expect(find.text('Caja 1'), findsOneWidget);
  });

  testWidgets('TestConsumosReporte_Empty_ShowsMessage', (tester) async {
    final repo = MockReportesRepo();
    when(() => repo.consumos(any())).thenAnswer(
      (_) async => const ConsumosReporte(total: 0, dias: [], cajeros: []),
    );
    await tester.pumpWidget(reportesHarness('/reportes/consumos', repo));
    await tester.pumpAndSettle();
    expect(find.text('No hubo consumos en este rango.'), findsOneWidget);
  });

  testWidgets('TestConsumosReporte_Error_RetryReloads', (tester) async {
    final repo = MockReportesRepo();
    var calls = 0;
    when(() => repo.consumos(any())).thenAnswer((_) async {
      if (calls++ == 0) throw const TimeoutFailure();
      return const ConsumosReporte(total: 0, dias: [], cajeros: []);
    });
    await tester.pumpWidget(reportesHarness('/reportes/consumos', repo));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(find.text('No hubo consumos en este rango.'), findsOneWidget);
  });

  testWidgets('TestReportes_TextScale200_NoOverflow', (tester) async {
    tester.view
      ..physicalSize = const Size(360, 640)
      ..devicePixelRatio = 1;
    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(() {
      tester.view.reset();
      tester.platformDispatcher.clearTextScaleFactorTestValue();
    });
    final repo = MockReportesRepo();
    when(repo.dashboard).thenAnswer((_) async => dashboardFixture);
    when(() => repo.list(any(), any()))
        .thenAnswer((_) async => pagina([reporteItem('c1')], total: 2));
    when(() => repo.consumos(any())).thenAnswer(
      (_) async => ConsumosReporte(
        total: 28,
        dias: dashboardFixture.consumos7d,
        cajeros: const [CajeroTotal(nombre: 'Caja 1', total: 28)],
      ),
    );
    for (final route in [
      '/reportes',
      '/reportes/sin-recarga',
      '/reportes/consumos',
    ]) {
      await tester.pumpWidget(reportesHarness(route, repo));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull, reason: route);
    }
  });
}
