import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';

AppFailure mapDioException(DioException e) {
  final inner = e.error;
  if (inner is AppFailure) return inner;
  switch (e.type) {
    case DioExceptionType.connectionTimeout:
    case DioExceptionType.sendTimeout:
    case DioExceptionType.receiveTimeout:
    case DioExceptionType.transformTimeout:
      return const TimeoutFailure();
    case DioExceptionType.connectionError:
      return const NetworkFailure();
    case DioExceptionType.badResponse:
      return _fromResponse(e.response);
    case DioExceptionType.cancel:
    case DioExceptionType.badCertificate:
    case DioExceptionType.unknown:
      return const UnknownFailure();
  }
}

AppFailure _fromResponse(Response<dynamic>? response) {
  final status = response?.statusCode ?? 0;
  final body = response?.data;
  final code = body is Map<String, dynamic> ? body['code'] : null;
  final detail = body is Map<String, dynamic> ? body['detail'] : null;
  final codeText = code is String ? code : null;
  if (status == 401) return UnauthorizedFailure(code: codeText);
  if (codeText == null) return const UnknownFailure();
  return ApiFailure(
    status: status,
    code: codeText,
    detail: detail is String ? detail : null,
  );
}
