import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';

import 'fakes.dart';

const _base = 'https://api.test';

({int status, Object? body}) _api(RequestOptions o) {
  final auth = o.headers['Authorization'];
  return auth == 'Bearer new-access'
      ? (status: 200, body: {'ok': true})
      : (status: 401, body: {'code': 'UNAUTHORIZED'});
}

({Dio dio, FakeAdapter api, FakeAdapter refresh, List<int> expired}) _setup(
  MemoryTokenStorage storage,
  Responder refreshResponder,
) {
  final api = FakeAdapter(_api);
  final refresh = FakeAdapter(refreshResponder);
  final expired = <int>[];
  final dio = buildApiClient(
    baseUrl: _base,
    storage: storage,
    onSessionExpired: () => expired.add(1),
    refreshClient: Dio(baseOptions(_base))..httpClientAdapter = refresh,
    retryClient: Dio(baseOptions(_base))..httpClientAdapter = api,
  )..httpClientAdapter = api;
  return (dio: dio, api: api, refresh: refresh, expired: expired);
}

void main() {
  test('TestAuthInterceptor_ConcurrentUnauthorized_RefreshesOnce', () async {
    final storage = MemoryTokenStorage(
      const Tokens(access: 'old-access', refresh: 'old-refresh'),
    );
    final s = _setup(
      storage,
      (_) => (
        status: 200,
        body: {'access_token': 'new-access', 'refresh_token': 'new-refresh'},
      ),
    );

    final results = await Future.wait(
      List.generate(5, (_) => s.dio.get<dynamic>('/clientes')),
    );

    expect(results.map((r) => r.statusCode), everyElement(200));
    expect(s.refresh.requests, hasLength(1));
    expect(s.refresh.requests.single.data, {'refresh_token': 'old-refresh'});
    expect(storage.tokens?.refresh, 'new-refresh');
    expect(s.expired, isEmpty);
  });

  test('TestAuthInterceptor_RefreshRejected_ClearsSessionAndFails', () async {
    final storage = MemoryTokenStorage(
      const Tokens(access: 'old-access', refresh: 'reused'),
    );
    final s = _setup(
      storage,
      (_) => (status: 401, body: {'code': 'INVALID_REFRESH_TOKEN'}),
    );

    await expectLater(
      s.dio.get<dynamic>('/clientes'),
      throwsA(
        isA<DioException>().having(
          (e) => e.error,
          'error',
          isA<UnauthorizedFailure>(),
        ),
      ),
    );
    expect(storage.tokens, isNull);
    expect(s.expired, hasLength(1));
  });

  test('TestAuthInterceptor_LoginPath_NoBearerNoRefresh', () async {
    final storage = MemoryTokenStorage(
      const Tokens(access: 'old-access', refresh: 'r'),
    );
    final s = _setup(storage, (_) => (status: 500, body: null));

    await expectLater(
      s.dio.post<dynamic>('/auth/login'),
      throwsA(isA<DioException>()),
    );
    expect(s.api.requests.single.headers['Authorization'], isNull);
    expect(s.refresh.requests, isEmpty);
  });
}
