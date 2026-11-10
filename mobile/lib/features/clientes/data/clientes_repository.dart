import 'package:dio/dio.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/network/error_mapper.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';

class ClientesRepository {
  new(this._dio);

  static const int pageSize = 50;

  final Dio _dio;

  Future<ClienteCreado> create(NuevoCliente input) async {
    final data = await _send('POST', '/clientes', data: input.toJson());
    try {
      return ClienteCreado(
        cliente: Cliente.fromJson(data['cliente'] as Map<String, dynamic>),
        tiquetera: Tiquetera.fromJson(
          data['tiquetera'] as Map<String, dynamic>,
        ),
        qrPayload: data['qr_payload'] as String,
      );
    } on Object {
      throw const UnknownFailure();
    }
  }

  Future<ClientePagina> list({String query = '', int page = 1}) async {
    final data = await _send(
      'GET',
      '/clientes',
      query: {
        if (query.trim().isNotEmpty) 'q': query.trim(),
        'page': page,
        'size': pageSize,
      },
    );
    try {
      return ClientePagina(
        items: (data['items'] as List<dynamic>)
            .map((e) => Cliente.fromJson(e as Map<String, dynamic>))
            .toList(growable: false),
        page: data['page'] as int,
        total: data['total'] as int,
      );
    } on Object {
      throw const UnknownFailure();
    }
  }

  Future<Cliente> get(String id) async {
    final data = await _send('GET', _p(id));
    return _parse<Cliente>(() => Cliente.fromJson(data));
  }

  Future<Tiquetera> recargar(
    String id, {
    required int almuerzos,
    required DateTime fechaVencimiento,
  }) async {
    final data = await _send(
      'POST',
      '${_p(id)}/tiqueteras',
      data: {
        'almuerzos': almuerzos,
        'fecha_vencimiento': formatApiDate(fechaVencimiento),
      },
    );
    return _parse<Tiquetera>(() => Tiquetera.fromJson(data));
  }

  Future<String> regenerarQr(String id) async {
    final data = await _send('POST', '${_p(id)}/qr/regenerar');
    final payload = data['qr_payload'];
    if (payload is! String) throw const UnknownFailure();
    return payload;
  }

  String _p(String id) => '/clientes/${Uri.encodeComponent(id)}';

  T _parse<T>(T Function() fn) {
    try {
      return fn();
    } on Object {
      throw const UnknownFailure();
    }
  }

  Future<Map<String, dynamic>> _send(
    String method,
    String path, {
    Object? data,
    Map<String, dynamic>? query,
  }) async {
    try {
      final res = await _dio.request<Map<String, dynamic>>(
        path,
        data: data,
        queryParameters: query,
        options: Options(method: method),
      );
      final body = res.data;
      if (body == null) throw const UnknownFailure();
      return body;
    } on DioException catch (e) {
      throw mapDioException(e);
    }
  }
}
