// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_User _$UserFromJson(Map<String, dynamic> json) => _User(
  id: json['id'] as String,
  nombre: json['nombre'] as String,
  rol: $enumDecodeNullable(
    _$RoleEnumMap,
    json['rol'],
    unknownValue: JsonKey.nullForUndefinedEnumValue,
  ),
);

Map<String, dynamic> _$UserToJson(_User instance) => <String, dynamic>{
  'id': instance.id,
  'nombre': instance.nombre,
  'rol': _$RoleEnumMap[instance.rol],
};

const _$RoleEnumMap = {Role.admin: 'admin', Role.cajero: 'cajero'};
