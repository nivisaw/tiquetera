import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/core/config/env.dart';
import 'package:tiquetera/core/network/dio_client.dart';
import 'package:tiquetera/core/storage/token_storage.dart';

part 'providers.g.dart';

@Riverpod(keepAlive: true)
Env env(Ref ref) => Env.fromDartDefines();

@Riverpod(keepAlive: true)
TokenStorage tokenStorage(Ref ref) =>
    SecureTokenStorage(const FlutterSecureStorage());

@Riverpod(keepAlive: true)
class SessionExpired extends _$SessionExpired {
  @override
  int build() => 0;

  void notify() => state++;
}

@Riverpod(keepAlive: true)
Dio apiClient(Ref ref) => buildApiClient(
  baseUrl: ref.watch(envProvider).apiBaseUrl,
  storage: ref.watch(tokenStorageProvider),
  onSessionExpired: ref.read(sessionExpiredProvider.notifier).notify,
);
