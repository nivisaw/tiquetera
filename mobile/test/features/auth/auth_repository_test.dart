import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';
import 'package:tiquetera/features/auth/data/auth_repository.dart';
import 'package:tiquetera/features/auth/domain/user.dart';

import '../../core/network/fakes.dart';

const _base = 'https://api.test';

({AuthRepository repo, FakeAdapter api, MemoryTokenStorage storage}) _setup(
  Responder responder, [
  Tokens? tokens,
]) {
  final storage = MemoryTokenStorage(tokens);
  final api = FakeAdapter(responder);
  final dio = buildApiClient(
    baseUrl: _base,
    storage: storage,
    onSessionExpired: () {},
    refreshClient: Dio(baseOptions(_base))..httpClientAdapter = api,
    retryClient: Dio(baseOptions(_base))..httpClientAdapter = api,
  )..httpClientAdapter = api;
  return (repo: AuthRepository(dio, storage), api: api, storage: storage);
}

const Map<String, Object> _loginOk = {
  'access_token': 'a1',
  'refresh_token': 'r1',
  'expires_in': 900,
  'user': {'id': 'u1', 'nombre': 'Ana', 'rol': 'admin'},
};

void main() {
  test('TestAuthRepository_LoginOk_PersistsSessionAndRestores', () async {
    final s = _setup((_) => (status: 200, body: _loginOk));

    final user = await s.repo.login(username: 'ana', password: 'secreto-123');

    expect(user.rol, Role.admin);
    expect(s.storage.tokens?.refresh, 'r1');
    expect(await s.repo.restore(), user);
  });

  test('TestAuthRepository_LoginInvalid_ThrowsUnauthorized', () async {
    final s = _setup(
      (_) => (status: 401, body: {'code': 'INVALID_CREDENTIALS'}),
    );

    await expectLater(
      s.repo.login(username: 'ana', password: 'mala-clave'),
      throwsA(isA<UnauthorizedFailure>()),
    );
    expect(s.storage.tokens, isNull);
  });

  test('TestAuthRepository_LoginUnknownRole_ThrowsAndKeepsNoSession', () async {
    final s = _setup(
      (_) => (
        status: 200,
        body: {
          ..._loginOk,
          'user': {'id': 'u1', 'nombre': 'Ana', 'rol': 'root'},
        },
      ),
    );

    await expectLater(
      s.repo.login(username: 'ana', password: 'secreto-123'),
      throwsA(isA<UnknownFailure>()),
    );
    expect(s.storage.tokens, isNull);
  });

  test('TestAuthRepository_LogoutOffline_ClearsLocalSession', () async {
    final s = _setup(
      (o) => o.path == '/auth/logout'
          ? throw DioException.connectionError(
              requestOptions: o,
              reason: 'offline',
            )
          : (status: 200, body: _loginOk),
    );
    await s.repo.login(username: 'ana', password: 'secreto-123');

    await s.repo.logout();

    expect(s.storage.tokens, isNull);
    expect(await s.repo.restore(), isNull);
  });

  test('TestAuthRepository_Logout_SendsRefreshToken', () async {
    final s = _setup(
      (o) => o.path == '/auth/logout'
          ? (status: 204, body: null)
          : (status: 200, body: _loginOk),
    );
    await s.repo.login(username: 'ana', password: 'secreto-123');

    await s.repo.logout();

    final req = s.api.requests.last;
    expect(req.path, '/auth/logout');
    expect(req.data, {'refresh_token': 'r1'});
    expect(req.headers['Authorization'], 'Bearer a1');
  });
}
