import 'package:freezed_annotation/freezed_annotation.dart';

part 'cliente.freezed.dart';
part 'cliente.g.dart';

@freezed
abstract class Tiquetera with _$Tiquetera {
  const factory({
    required String id,
    @JsonKey(name: 'almuerzos_comprados') required int almuerzosComprados,
    required int saldo,
    @JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate)
    required DateTime fechaVencimiento,
    required bool activa,
  }) = _Tiquetera;

  factory fromJson(Map<String, dynamic> json) => _$TiqueteraFromJson(json);
}

/// Cliente; `tiquetera` solo viene en el detalle (`GET /clientes/{id}`).
@freezed
abstract class Cliente with _$Cliente {
  const factory({
    required String id,
    required String nombre,
    required String telefono,
    required bool activo,
    String? documento,
    String? email,
    Tiquetera? tiquetera,
  }) = _Cliente;

  factory fromJson(Map<String, dynamic> json) => _$ClienteFromJson(json);
}

/// Resultado de crear un cliente: único momento (con regenerar) con QR.
class ClienteCreado {
  const new({
    required this.cliente,
    required this.tiquetera,
    required this.qrPayload,
  });

  final Cliente cliente;
  final Tiquetera tiquetera;
  final String qrPayload;
}

class ClientePagina {
  const new({required this.items, required this.page, required this.total});

  final List<Cliente> items;
  final int page;
  final int total;
}

class NuevoCliente {
  const new({
    required this.nombre,
    required this.telefono,
    required this.almuerzos,
    required this.fechaVencimiento,
    this.documento,
    this.email,
  });

  final String nombre;
  final String telefono;
  final String? documento;
  final String? email;
  final int almuerzos;
  final DateTime fechaVencimiento;

  Map<String, dynamic> toJson() => {
    'nombre': nombre,
    'telefono': telefono,
    'documento': ?documento,
    'email': ?email,
    'almuerzos': almuerzos,
    'fecha_vencimiento': formatApiDate(fechaVencimiento),
  };
}

DateTime parseApiDate(String value) {
  final d = DateTime.parse(value);
  return DateTime(d.year, d.month, d.day);
}

String formatApiDate(DateTime d) =>
    '${d.year.toString().padLeft(4, '0')}-'
    '${d.month.toString().padLeft(2, '0')}-'
    '${d.day.toString().padLeft(2, '0')}';
