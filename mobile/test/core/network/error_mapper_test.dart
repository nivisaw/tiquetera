import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/error_mapper.dart';

DioException _bad(int status, Object? body) {
  final req = RequestOptions(path: '/x');
  return DioException(
    requestOptions: req,
    type: DioExceptionType.badResponse,
    response: Response<dynamic>(
      requestOptions: req,
      statusCode: status,
      data: body,
    ),
  );
}

void main() {
  test('TestMapDioException_Problem422_ApiFailureWithCode', () {
    final f = mapDioException(
      _bad(422, {
        'type': 'about:blank',
        'status': 422,
        'code': 'TIQUETERA_VENCIDA',
        'detail': 'Vencida',
      }),
    );
    expect(
      f,
      isA<ApiFailure>()
          .having((f) => f.status, 'status', 422)
          .having((f) => f.code, 'code', 'TIQUETERA_VENCIDA')
          .having((f) => f.detail, 'detail', 'Vencida'),
    );
  });

  test('TestMapDioException_401_Unauthorized', () {
    expect(
      mapDioException(_bad(401, {'code': 'INVALID_CREDENTIALS'})),
      isA<UnauthorizedFailure>().having(
        (f) => f.code,
        'code',
        'INVALID_CREDENTIALS',
      ),
    );
  });

  test('TestMapDioException_BodyWithoutCode_Unknown', () {
    expect(mapDioException(_bad(502, 'Bad gateway')), isA<UnknownFailure>());
  });

  test('TestMapDioException_TransportErrors_Mapped', () {
    final req = RequestOptions(path: '/x');
    DioException of(DioExceptionType t) =>
        DioException(requestOptions: req, type: t);
    expect(
      mapDioException(of(DioExceptionType.receiveTimeout)),
      isA<TimeoutFailure>(),
    );
    expect(
      mapDioException(of(DioExceptionType.connectionError)),
      isA<NetworkFailure>(),
    );
  });
}
