import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';
import 'package:tiquetera/features/clientes/data/clientes_repository.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';

import '../../core/network/fakes.dart';
import 'fixtures.dart';

const _base = 'https://api.test';

({ClientesRepository repo, FakeAdapter api}) _setup(Responder responder) {
  final storage = MemoryTokenStorage(const Tokens(access: 'a', refresh: 'r'));
  final api = FakeAdapter(responder);
  final dio = buildApiClient(
    baseUrl: _base,
    storage: storage,
    onSessionExpired: () {},
    refreshClient: Dio(baseOptions(_base))..httpClientAdapter = api,
    retryClient: Dio(baseOptions(_base))..httpClientAdapter = api,
  )..httpClientAdapter = api;
  return (repo: ClientesRepository(dio), api: api);
}

void main() {
  test('TestClientesRepository_Create_SendsBodyAndParses', () async {
    final s = _setup(
      (_) => (
        status: 201,
        body: {
          'cliente': clienteJson,
          'tiquetera': tiqueteraJson,
          'qr_payload': 'TQ1.abc',
        },
      ),
    );

    final r = await s.repo.create(
      NuevoCliente(
        nombre: 'Ana Pérez',
        telefono: '+573001234567',
        email: 'ana@x.co',
        almuerzos: 20,
        fechaVencimiento: DateTime(2026, 12, 10),
      ),
    );

    final req = s.api.requests.single;
    expect(req.method, 'POST');
    expect(req.path, '/clientes');
    expect(req.data, containsPair('fecha_vencimiento', '2026-12-10'));
    expect(req.data, isNot(contains('documento')));
    expect(r.qrPayload, 'TQ1.abc');
    expect(r.tiquetera.saldo, 20);
    expect(r.tiquetera.fechaVencimiento, DateTime(2026, 12, 10));
  });

  test('TestClientesRepository_CreateDuplicado_ThrowsApiFailure', () async {
    final s = _setup(
      (_) => (status: 409, body: {'status': 409, 'code': 'TELEFONO_DUPLICADO'}),
    );
    await expectLater(
      s.repo.create(
        NuevoCliente(
          nombre: 'Ana',
          telefono: '+573001234567',
          almuerzos: 1,
          fechaVencimiento: DateTime(2026, 12, 10),
        ),
      ),
      throwsA(
        isA<ApiFailure>().having((f) => f.code, 'code', 'TELEFONO_DUPLICADO'),
      ),
    );
  });

  test('TestClientesRepository_List_SendsQueryAndParses', () async {
    final s = _setup(
      (_) => (
        status: 200,
        body: {
          'items': [clienteJson],
          'page': 2,
          'size': 50,
          'total': 51,
        },
      ),
    );

    final page = await s.repo.list(query: ' ana ', page: 2);

    expect(s.api.requests.single.queryParameters, {
      'q': 'ana',
      'page': 2,
      'size': 50,
    });
    expect(page.items.single.nombre, 'Ana Pérez');
    expect(page.items.single.tiquetera, isNull);
    expect(page.total, 51);
  });

  test('TestClientesRepository_GetRecargarRegenerar_UsesRoutes', () async {
    final s = _setup((o) {
      if (o.path.endsWith('/tiqueteras')) {
        return (status: 201, body: tiqueteraJson);
      }
      if (o.path.endsWith('/qr/regenerar')) {
        return (status: 200, body: {'qr_payload': 'TQ1.new'});
      }
      return (status: 200, body: {...clienteJson, 'tiquetera': tiqueteraJson});
    });

    final c = await s.repo.get('c1');
    final t = await s.repo.recargar(
      'c1',
      almuerzos: 10,
      fechaVencimiento: DateTime(2027, 1, 2),
    );
    final qr = await s.repo.regenerarQr('c1');

    expect(c.tiquetera?.saldo, 20);
    expect(t.id, 't1');
    expect(qr, 'TQ1.new');
    expect(s.api.requests.map((r) => '${r.method} ${r.path}'), [
      'GET /clientes/c1',
      'POST /clientes/c1/tiqueteras',
      'POST /clientes/c1/qr/regenerar',
    ]);
    expect(s.api.requests[1].data, {
      'almuerzos': 10,
      'fecha_vencimiento': '2027-01-02',
    });
  });

  test('TestClientesRepository_MalformedBody_ThrowsUnknown', () async {
    final s = _setup((_) => (status: 200, body: {'items': 'x'}));
    await expectLater(s.repo.list(), throwsA(isA<UnknownFailure>()));
    final r = _setup((_) => (status: 200, body: {'nope': 1}));
    await expectLater(r.repo.regenerarQr('c1'), throwsA(isA<UnknownFailure>()));
    await expectLater(r.repo.get('c1'), throwsA(isA<UnknownFailure>()));
  });
}
