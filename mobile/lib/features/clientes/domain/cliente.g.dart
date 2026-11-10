// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'cliente.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Tiquetera _$TiqueteraFromJson(Map<String, dynamic> json) => _Tiquetera(
  id: json['id'] as String,
  almuerzosComprados: (json['almuerzos_comprados'] as num).toInt(),
  saldo: (json['saldo'] as num).toInt(),
  fechaVencimiento: parseApiDate(json['fecha_vencimiento'] as String),
  activa: json['activa'] as bool,
);

Map<String, dynamic> _$TiqueteraToJson(_Tiquetera instance) =>
    <String, dynamic>{
      'id': instance.id,
      'almuerzos_comprados': instance.almuerzosComprados,
      'saldo': instance.saldo,
      'fecha_vencimiento': instance.fechaVencimiento.toIso8601String(),
      'activa': instance.activa,
    };

_Cliente _$ClienteFromJson(Map<String, dynamic> json) => _Cliente(
  id: json['id'] as String,
  nombre: json['nombre'] as String,
  telefono: json['telefono'] as String,
  activo: json['activo'] as bool,
  documento: json['documento'] as String?,
  email: json['email'] as String?,
  tiquetera: json['tiquetera'] == null
      ? null
      : Tiquetera.fromJson(json['tiquetera'] as Map<String, dynamic>),
);

Map<String, dynamic> _$ClienteToJson(_Cliente instance) => <String, dynamic>{
  'id': instance.id,
  'nombre': instance.nombre,
  'telefono': instance.telefono,
  'activo': instance.activo,
  'documento': instance.documento,
  'email': instance.email,
  'tiquetera': instance.tiquetera,
};
