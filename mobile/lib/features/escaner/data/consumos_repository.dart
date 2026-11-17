import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/error_mapper.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';

class ConsumosRepository {
  new(this._dio);

  final Dio _dio;

  Future<ConsumoOk> registrar({
    required String qr,
    required String idempotencyKey,
  }) async {
    final Map<String, dynamic>? body;
    try {
      final res = await _dio.post<Map<String, dynamic>>(
        '/consumos',
        data: {'qr': qr},
        options: Options(headers: {'Idempotency-Key': idempotencyKey}),
      );
      body = res.data;
    } on DioException catch (e) {
      throw mapDioException(e);
    }
    try {
      return ConsumoOk.fromJson(body!);
    } on Object {
      throw const UnknownFailure();
    }
  }
}
