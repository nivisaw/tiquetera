import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/auth_interceptor.dart';
import 'package:tiquetera/core/network/error_mapper.dart';
import 'package:tiquetera/core/storage/token_storage.dart';

const _timeout = Duration(seconds: 10);

BaseOptions baseOptions(String baseUrl) => BaseOptions(
  baseUrl: baseUrl,
  connectTimeout: _timeout,
  sendTimeout: _timeout,
  receiveTimeout: _timeout,
  contentType: Headers.jsonContentType,
);

final class ErrorMappingInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    handler.next(err.copyWith(error: mapDioException(err)));
  }
}

Dio buildApiClient({
  required String baseUrl,
  required TokenStorage storage,
  required void Function() onSessionExpired,
  Dio? refreshClient,
  Dio? retryClient,
}) {
  final dio = Dio(baseOptions(baseUrl));
  dio.interceptors.addAll([
    AuthInterceptor(
      storage: storage,
      refreshClient: refreshClient ?? Dio(baseOptions(baseUrl)),
      retryClient: retryClient ?? Dio(baseOptions(baseUrl)),
      onSessionExpired: onSessionExpired,
    ),
    ErrorMappingInterceptor(),
  ]);
  return dio;
}
