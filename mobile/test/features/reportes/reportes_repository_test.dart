import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';
import 'package:tiquetera/features/reportes/data/reportes_repository.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';

import '../../core/network/fakes.dart';

const _base = 'https://api.test';

({ReportesRepository repo, FakeAdapter api}) _setup(Responder responder) {
  final storage = MemoryTokenStorage(const Tokens(access: 'a', refresh: 'r'));
  final api = FakeAdapter(responder);
  final dio = buildApiClient(
    baseUrl: _base,
    storage: storage,
    onSessionExpired: () {},
    refreshClient: Dio(baseOptions(_base))..httpClientAdapter = api,
    retryClient: Dio(baseOptions(_base))..httpClientAdapter = api,
  )..httpClientAdapter = api;
  return (repo: ReportesRepository(dio), api: api);
}

void main() {
  test('TestReportesRepository_Dashboard_Parses', () async {
    final s = _setup(
      (_) => (
        status: 200,
        body: {
          'consumos_hoy': 3,
          'consumos_7d': [
            {'fecha': '2026-11-04', 'total': 3},
          ],
          'clientes_vigentes': 4,
          'saldo_pendiente': 40,
          'vendidos_mes': 60,
          'por_vencer_7d': 1,
          'rechazos_hoy': 2,
        },
      ),
    );

    final d = await s.repo.dashboard();

    expect(s.api.requests.single.path, '/reportes/dashboard');
    expect(d.consumos7d.single.fecha, DateTime(2026, 11, 4));
    expect(d.saldoPendiente, 40);
    expect(d.rechazosHoy, 2);
  });

  test(
    'TestReportesRepository_SinRecarga_SendsParamsAndUsesBogotaDate',
    () async {
      final s = _setup(
        (_) => (
          status: 200,
          body: {
            'items': [
              {
                'cliente_id': 'c1',
                'nombre': 'Ana',
                'telefono': '+573001234567',
                'ultima_recarga': '2026-10-05T03:00:00Z',
                'saldo': 2,
              },
            ],
            'page': 2,
            'size': 50,
            'total': 51,
          },
        ),
      );

      final p = await s.repo.list(ReporteTipo.sinRecarga, 60, page: 2);

      final req = s.api.requests.single;
      expect(req.path, '/reportes/sin-recarga');
      expect(req.queryParameters, {'dias': 60, 'page': 2, 'size': 50});
      expect(p.items.single.fecha, DateTime(2026, 10, 4));
      expect(p.total, 51);
    },
  );

  test('TestReportesRepository_SaldoBajo_ParsesExpiry', () async {
    final s = _setup(
      (_) => (
        status: 200,
        body: {
          'items': [
            {
              'cliente_id': 'c1',
              'nombre': 'Ana',
              'telefono': '+573001234567',
              'fecha_vencimiento': '2026-12-01',
              'saldo': 1,
            },
          ],
          'page': 1,
          'size': 50,
          'total': 1,
        },
      ),
    );

    final p = await s.repo.list(ReporteTipo.saldoBajo, 3);

    expect(s.api.requests.single.queryParameters['max'], 3);
    expect(p.items.single.fecha, DateTime(2026, 12));
  });

  test('TestReportesRepository_Consumos_SendsRange', () async {
    final s = _setup(
      (_) => (
        status: 200,
        body: {
          'total': 5,
          'dias': [
            {'fecha': '2026-11-01', 'total': 5},
          ],
          'cajeros': [
            {'cajero_id': 'u1', 'nombre': 'Caja 1', 'total': 5},
          ],
        },
      ),
    );

    final r = await s.repo.consumos(
      RangoFechas(DateTime(2026, 11), DateTime(2026, 11, 7)),
    );

    expect(s.api.requests.single.queryParameters, {
      'desde': '2026-11-01',
      'hasta': '2026-11-07',
    });
    expect(r.cajeros.single.nombre, 'Caja 1');
  });

  test('TestReportesRepository_Malformed_ThrowsUnknown', () async {
    final s = _setup((_) => (status: 200, body: {'x': 1}));
    expect(s.repo.dashboard(), throwsA(isA<UnknownFailure>()));
  });

  test('TestReportesRepository_ServerError_MapsFailure', () async {
    final s = _setup(
      (_) => (
        status: 422,
        body: {'code': 'VALIDATION_ERROR', 'title': 'x', 'status': 422},
      ),
    );
    expect(s.repo.list(ReporteTipo.porVencer, 7), throwsA(isA<ApiFailure>()));
  });

  test('TestRangoFechas_Valido_Max31Days', () {
    final d = DateTime(2026, 11);
    expect(RangoFechas(d, d.add(const Duration(days: 30))).valido, isTrue);
    expect(RangoFechas(d, d.add(const Duration(days: 31))).valido, isFalse);
    expect(RangoFechas(d, d.subtract(const Duration(days: 1))).valido, isFalse);
    expect(
      RangoFechas.ultimaSemana(DateTime(2026, 11, 7)),
      RangoFechas(d, DateTime(2026, 11, 7)),
    );
    expect(ReporteTipo.fromPath('nada'), isNull);
  });
}
