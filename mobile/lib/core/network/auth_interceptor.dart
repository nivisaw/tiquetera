import 'dart:async';

import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/storage/token_storage.dart';

const _retriedKey = 'auth_retried';
const _publicPaths = {'/auth/login', '/auth/refresh'};

/// Adjunta el access token y, ante un 401, refresca una sola vez aunque
/// varias peticiones fallen a la vez (las demás esperan el mismo refresh).
class AuthInterceptor extends Interceptor {
  new({
    required this._storage,
    required this._refreshClient,
    required this._retryClient,
    required this._onSessionExpired,
  });

  final TokenStorage _storage;
  final Dio _refreshClient;
  final Dio _retryClient;
  final void Function() _onSessionExpired;
  Future<Tokens?>? _inFlightRefresh;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    if (!_publicPaths.contains(options.path)) {
      final tokens = await _storage.read();
      if (tokens != null) {
        options.headers['Authorization'] = 'Bearer ${tokens.access}';
      }
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final request = err.requestOptions;
    final canRetry =
        err.response?.statusCode == 401 &&
        !_publicPaths.contains(request.path) &&
        request.extra[_retriedKey] != true;
    if (!canRetry) return handler.next(err);

    final Tokens? tokens;
    try {
      tokens = await _refreshOnce();
    } on DioException catch (e) {
      return handler.next(e);
    }
    if (tokens == null) {
      return handler.next(err.copyWith(error: const UnauthorizedFailure()));
    }
    request.extra[_retriedKey] = true;
    request.headers['Authorization'] = 'Bearer ${tokens.access}';
    try {
      handler.resolve(await _retryClient.fetch<dynamic>(request));
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  Future<Tokens?> _refreshOnce() {
    return _inFlightRefresh ??= _refresh().whenComplete(
      () => _inFlightRefresh = null,
    );
  }

  Future<Tokens?> _refresh() async {
    final current = await _storage.read();
    if (current == null) return await _expire();
    try {
      final res = await _refreshClient.post<Map<String, dynamic>>(
        '/auth/refresh',
        data: {'refresh_token': current.refresh},
      );
      final access = res.data?['access_token'];
      final refresh = res.data?['refresh_token'];
      if (access is! String || refresh is! String) return await _expire();
      final tokens = Tokens(access: access, refresh: refresh);
      await _storage.write(tokens);
      return tokens;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) return await _expire();
      rethrow;
    }
  }

  Future<Tokens?> _expire() async {
    await _storage.clear();
    _onSessionExpired();
    return null;
  }
}
