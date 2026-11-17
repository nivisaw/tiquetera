/// Débito aceptado por el servidor.
class ConsumoOk {
  const new({
    required this.nombre,
    required this.saldoRestante,
    required this.fechaVencimiento,
  });

  factory fromJson(Map<String, dynamic> json) => ConsumoOk(
    nombre: (json['cliente'] as Map<String, dynamic>)['nombre'] as String,
    saldoRestante: json['saldo_restante'] as int,
    fechaVencimiento: DateTime.parse(json['fecha_vencimiento'] as String),
  );

  final String nombre;
  final int saldoRestante;
  final DateTime fechaVencimiento;
}

/// Motivo de rechazo mostrado al cajero.
enum RechazoMotivo {
  qrInvalido,
  clienteNoExiste,
  clienteInactivo,
  sinTiquetera,
  tiqueteraVencida,
  sinSaldo,
  sinRed,
  demasiadosIntentos,
  error;

  /// Mapea el `code` de un 422 de `POST /consumos`.
  static RechazoMotivo fromCode(String code) => switch (code) {
    'QR_INVALIDO' => qrInvalido,
    'CLIENTE_NO_EXISTE' => clienteNoExiste,
    'CLIENTE_INACTIVO' => clienteInactivo,
    'SIN_TIQUETERA' => sinTiquetera,
    'TIQUETERA_VENCIDA' => tiqueteraVencida,
    'SIN_SALDO' => sinSaldo,
    _ => error,
  };
}
