import 'package:flutter/foundation.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';

/// Contrato: `docs/API.md#reportes-admin`.
class DiaTotal {
  const new({required this.fecha, required this.total});

  factory fromJson(Map<String, dynamic> json) => DiaTotal(
    fecha: parseApiDate(json['fecha'] as String),
    total: json['total'] as int,
  );

  final DateTime fecha;
  final int total;
}

class Dashboard {
  const new({
    required this.consumosHoy,
    required this.consumos7d,
    required this.clientesVigentes,
    required this.saldoPendiente,
    required this.vendidosMes,
    required this.porVencer7d,
    required this.rechazosHoy,
  });

  factory fromJson(Map<String, dynamic> json) => Dashboard(
    consumosHoy: json['consumos_hoy'] as int,
    consumos7d: _list(json['consumos_7d'], DiaTotal.fromJson),
    clientesVigentes: json['clientes_vigentes'] as int,
    saldoPendiente: json['saldo_pendiente'] as int,
    vendidosMes: json['vendidos_mes'] as int,
    porVencer7d: json['por_vencer_7d'] as int,
    rechazosHoy: json['rechazos_hoy'] as int,
  );

  final int consumosHoy;
  final List<DiaTotal> consumos7d;
  final int clientesVigentes;
  final int saldoPendiente;
  final int vendidosMes;
  final int porVencer7d;
  final int rechazosHoy;
}

/// Listados de clientes del admin, con sus valores de filtro sugeridos.
enum ReporteTipo {
  sinRecarga('sin-recarga', 'dias', 30, [15, 30, 60, 90]),
  porVencer('por-vencer', 'dias', 7, [3, 7, 15, 30]),
  saldoBajo('saldo-bajo', 'max', 3, [1, 3, 5, 10]);

  new(this.path, this.param, this.defaultValue, this.opciones);

  final String path;
  final String param;
  final int defaultValue;
  final List<int> opciones;

  static ReporteTipo? fromPath(String path) =>
      values.where((t) => t.path == path).firstOrNull;
}

/// Fila de un listado; `fecha` es la última recarga (sin-recarga) o el
/// vencimiento (por-vencer, saldo-bajo), ya como fecha de Bogotá.
class ClienteReporte {
  const new({
    required this.clienteId,
    required this.nombre,
    required this.telefono,
    required this.fecha,
    required this.saldo,
  });

  factory fromJson(ReporteTipo tipo, Map<String, dynamic> json) =>
      ClienteReporte(
        clienteId: json['cliente_id'] as String,
        nombre: json['nombre'] as String,
        telefono: json['telefono'] as String,
        fecha: tipo == ReporteTipo.sinRecarga
            ? ClienteRules.todayBogota(
                DateTime.parse(json['ultima_recarga'] as String),
              )
            : parseApiDate(json['fecha_vencimiento'] as String),
        saldo: json['saldo'] as int,
      );

  final String clienteId;
  final String nombre;
  final String telefono;
  final DateTime fecha;
  final int saldo;
}

class ReportePagina {
  const new({required this.items, required this.page, required this.total});

  final List<ClienteReporte> items;
  final int page;
  final int total;
}

class CajeroTotal {
  const new({required this.nombre, required this.total});

  factory fromJson(Map<String, dynamic> json) => CajeroTotal(
    nombre: json['nombre'] as String,
    total: json['total'] as int,
  );

  final String nombre;
  final int total;
}

class ConsumosReporte {
  const new({required this.total, required this.dias, required this.cajeros});

  factory fromJson(Map<String, dynamic> json) => ConsumosReporte(
    total: json['total'] as int,
    dias: _list(json['dias'], DiaTotal.fromJson),
    cajeros: _list(json['cajeros'], CajeroTotal.fromJson),
  );

  final int total;
  final List<DiaTotal> dias;
  final List<CajeroTotal> cajeros;
}

/// Rango del reporte de consumos (fechas de Bogotá, inclusivo).
@immutable
class RangoFechas {
  const new(this.desde, this.hasta);

  /// Últimos 7 días terminando en [hoy].
  factory ultimaSemana(DateTime hoy) =>
      RangoFechas(hoy.subtract(const Duration(days: 6)), hoy);

  static const int maxDias = 31;

  final DateTime desde;
  final DateTime hasta;

  bool get valido =>
      !hasta.isBefore(desde) && hasta.difference(desde).inDays < maxDias;

  @override
  bool operator ==(Object other) =>
      other is RangoFechas && other.desde == desde && other.hasta == hasta;

  @override
  int get hashCode => Object.hash(desde, hasta);
}

List<T> _list<T>(Object? raw, T Function(Map<String, dynamic>) fromJson) =>
    (raw! as List<dynamic>)
        .map((e) => fromJson(e as Map<String, dynamic>))
        .toList(growable: false);
