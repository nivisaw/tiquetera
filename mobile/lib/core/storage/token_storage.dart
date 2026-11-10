import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class Tokens {
  const new({required this.access, required this.refresh});
  final String access;
  final String refresh;
}

abstract interface class TokenStorage {
  Future<Tokens?> read();
  Future<void> write(Tokens tokens);
  Future<String?> readUser();
  Future<void> writeUser(String userJson);
  Future<void> clear();
}

class SecureTokenStorage implements TokenStorage {
  new(this._storage);

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';
  static const _userKey = 'user';
  final FlutterSecureStorage _storage;

  @override
  Future<Tokens?> read() async {
    final access = await _storage.read(key: _accessKey);
    final refresh = await _storage.read(key: _refreshKey);
    if (access == null || refresh == null) return null;
    return Tokens(access: access, refresh: refresh);
  }

  @override
  Future<void> write(Tokens tokens) async {
    await _storage.write(key: _accessKey, value: tokens.access);
    await _storage.write(key: _refreshKey, value: tokens.refresh);
  }

  @override
  Future<String?> readUser() => _storage.read(key: _userKey);

  @override
  Future<void> writeUser(String userJson) =>
      _storage.write(key: _userKey, value: userJson);

  @override
  Future<void> clear() async {
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
    await _storage.delete(key: _userKey);
  }
}
