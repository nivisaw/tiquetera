import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/error_mapper.dart';
import 'package:tiquetera/core/storage/token_storage.dart';
import 'package:tiquetera/features/auth/domain/user.dart';

class AuthRepository {
  new(this._dio, this._storage);

  final Dio _dio;
  final TokenStorage _storage;

  Future<User> login({
    required String username,
    required String password,
  }) async {
    final Response<Map<String, dynamic>> res;
    try {
      res = await _dio.post<Map<String, dynamic>>(
        '/auth/login',
        data: {'username': username, 'password': password},
      );
    } on DioException catch (e) {
      throw mapDioException(e);
    }
    final data = res.data;
    final access = data?['access_token'];
    final refresh = data?['refresh_token'];
    final userJson = data?['user'];
    if (access is! String ||
        refresh is! String ||
        userJson is! Map<String, dynamic>) {
      throw const UnknownFailure();
    }
    final user = User.fromJson(userJson);
    if (user.rol == null) throw const UnknownFailure();
    await _storage.write(Tokens(access: access, refresh: refresh));
    await _storage.writeUser(jsonEncode(user.toJson()));
    return user;
  }

  Future<User?> restore() async {
    final tokens = await _storage.read();
    final raw = await _storage.readUser();
    if (tokens == null || raw == null) return null;
    try {
      final user = User.fromJson(jsonDecode(raw) as Map<String, dynamic>);
      return user.rol == null ? null : user;
    } on Object {
      await _storage.clear();
      return null;
    }
  }

  /// Revoca en el servidor si hay red; la sesión local se borra siempre.
  Future<void> logout() async {
    final tokens = await _storage.read();
    await _storage.clear();
    if (tokens == null) return;
    try {
      await _dio.post<void>(
        '/auth/logout',
        data: {'refresh_token': tokens.refresh},
        options: Options(headers: {'Authorization': 'Bearer ${tokens.access}'}),
      );
    } on DioException {
      return;
    }
  }
}
