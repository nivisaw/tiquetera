import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/error_mapper.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';

class ReportesRepository {
  new(this._dio);

  static const int pageSize = 50;

  final Dio _dio;

  Future<Dashboard> dashboard() async {
    final data = await _get('/reportes/dashboard');
    return _parse<Dashboard>(() => Dashboard.fromJson(data));
  }

  Future<ReportePagina> list(
    ReporteTipo tipo,
    int valor, {
    int page = 1,
  }) async {
    final data = await _get('/reportes/${tipo.path}', {
      tipo.param: valor,
      'page': page,
      'size': pageSize,
    });
    return _parse<ReportePagina>(
      () => ReportePagina(
        items: (data['items'] as List<dynamic>)
            .map(
              (e) => ClienteReporte.fromJson(tipo, e as Map<String, dynamic>),
            )
            .toList(growable: false),
        page: data['page'] as int,
        total: data['total'] as int,
      ),
    );
  }

  Future<ConsumosReporte> consumos(RangoFechas rango) async {
    final data = await _get('/reportes/consumos', {
      'desde': formatApiDate(rango.desde),
      'hasta': formatApiDate(rango.hasta),
    });
    return _parse<ConsumosReporte>(() => ConsumosReporte.fromJson(data));
  }

  T _parse<T>(T Function() fn) {
    try {
      return fn();
    } on Object {
      throw const UnknownFailure();
    }
  }

  Future<Map<String, dynamic>> _get(
    String path, [
    Map<String, dynamic>? query,
  ]) async {
    try {
      final res = await _dio.get<Map<String, dynamic>>(
        path,
        queryParameters: query,
      );
      final body = res.data;
      if (body == null) throw const UnknownFailure();
      return body;
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
