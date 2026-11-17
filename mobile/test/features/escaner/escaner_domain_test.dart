import 'dart:math';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';
import 'package:tiquetera/features/escaner/data/consumos_repository.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';
import 'package:tiquetera/features/escaner/domain/idempotency_key.dart';
import 'package:tiquetera/features/escaner/domain/scan_debouncer.dart';

import '../../core/network/fakes.dart';

const _base = 'https://api.test';

({ConsumosRepository repo, FakeAdapter api}) _setup(Responder responder) {
  final storage = MemoryTokenStorage(const Tokens(access: 'a', refresh: 'r'));
  final api = FakeAdapter(responder);
  final dio = buildApiClient(
    baseUrl: _base,
    storage: storage,
    onSessionExpired: () {},
    refreshClient: Dio(baseOptions(_base))..httpClientAdapter = api,
    retryClient: Dio(baseOptions(_base))..httpClientAdapter = api,
  )..httpClientAdapter = api;
  return (repo: ConsumosRepository(dio), api: api);
}

const Map<String, Object> okJson = {
  'ok': true,
  'cliente': {'nombre': 'Ana P.'},
  'saldo_restante': 7,
  'fecha_vencimiento': '2026-12-31',
  'fecha_hora': '2026-11-17T12:00:00Z',
};

void main() {
  group('ScanDebouncer', () {
    test('TestScanDebouncer_SameTokenWithin5s_Ignored', () {
      var now = DateTime(2026, 11, 17, 12);
      final d = ScanDebouncer(() => now);
      expect(d.accept('A'), isTrue);
      now = now.add(const Duration(seconds: 4));
      expect(d.accept('A'), isFalse);
      expect(d.accept('B'), isTrue);
      now = now.add(const Duration(seconds: 1));
      expect(d.accept('A'), isTrue);
    });

    test('TestScanDebouncer_Forget_AllowsImmediateRetry', () {
      final d = ScanDebouncer(() => DateTime(2026))
        ..accept('A')
        ..forget('A');
      expect(d.accept('A'), isTrue);
    });
  });

  test('TestIdempotencyKey_Generate_IsUuidV4AndUnique', () {
    final re = RegExp(
      r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
    );
    final a = newIdempotencyKey(Random(1));
    expect(re.hasMatch(a), isTrue);
    expect(re.hasMatch(newIdempotencyKey()), isTrue);
    expect(newIdempotencyKey(), isNot(newIdempotencyKey()));
  });

  test('TestRechazoMotivo_FromCode_MapsAllApiCodes', () {
    expect(RechazoMotivo.fromCode('QR_INVALIDO'), RechazoMotivo.qrInvalido);
    expect(
      RechazoMotivo.fromCode('CLIENTE_NO_EXISTE'),
      RechazoMotivo.clienteNoExiste,
    );
    expect(
      RechazoMotivo.fromCode('CLIENTE_INACTIVO'),
      RechazoMotivo.clienteInactivo,
    );
    expect(RechazoMotivo.fromCode('SIN_TIQUETERA'), RechazoMotivo.sinTiquetera);
    expect(
      RechazoMotivo.fromCode('TIQUETERA_VENCIDA'),
      RechazoMotivo.tiqueteraVencida,
    );
    expect(RechazoMotivo.fromCode('SIN_SALDO'), RechazoMotivo.sinSaldo);
    expect(RechazoMotivo.fromCode('OTRO'), RechazoMotivo.error);
  });

  group('ConsumosRepository', () {
    test('TestConsumosRepository_Registrar_SendsKeyAndParses', () async {
      final s = _setup((_) => (status: 200, body: okJson));
      final r = await s.repo.registrar(qr: 'TQ1.x', idempotencyKey: 'k-1');
      expect(r.nombre, 'Ana P.');
      expect(r.saldoRestante, 7);
      expect(r.fechaVencimiento, DateTime(2026, 12, 31));
      final req = s.api.requests.single;
      expect(req.method, 'POST');
      expect(req.path, '/consumos');
      expect(req.headers['Idempotency-Key'], 'k-1');
      expect(req.data, {'qr': 'TQ1.x'});
    });

    test('TestConsumosRepository_422_ThrowsApiFailureWithCode', () async {
      final s = _setup(
        (_) => (status: 422, body: {'code': 'SIN_SALDO', 'status': 422}),
      );
      await expectLater(
        s.repo.registrar(qr: 'TQ1.x', idempotencyKey: 'k'),
        throwsA(
          isA<ApiFailure>()
              .having((f) => f.status, 'status', 422)
              .having((f) => f.code, 'code', 'SIN_SALDO'),
        ),
      );
    });

    test('TestConsumosRepository_MalformedBody_ThrowsUnknown', () async {
      final s = _setup((_) => (status: 200, body: {'ok': true}));
      await expectLater(
        s.repo.registrar(qr: 'TQ1.x', idempotencyKey: 'k'),
        throwsA(isA<UnknownFailure>()),
      );
    });
  });
}
