// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'cliente.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Tiquetera {

 String get id;@JsonKey(name: 'almuerzos_comprados') int get almuerzosComprados; int get saldo;@JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate) DateTime get fechaVencimiento; bool get activa;
/// Create a copy of Tiquetera
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TiqueteraCopyWith<Tiquetera> get copyWith => _$TiqueteraCopyWithImpl<Tiquetera>(this as Tiquetera, _$identity);

  /// Serializes this Tiquetera to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Tiquetera;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Tiquetera&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.almuerzosComprados, _this.almuerzosComprados) || other.almuerzosComprados == _this.almuerzosComprados)&&(identical(other.saldo, _this.saldo) || other.saldo == _this.saldo)&&(identical(other.fechaVencimiento, _this.fechaVencimiento) || other.fechaVencimiento == _this.fechaVencimiento)&&(identical(other.activa, _this.activa) || other.activa == _this.activa));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Tiquetera;
  return Object.hash(runtimeType,_this.id,_this.almuerzosComprados,_this.saldo,_this.fechaVencimiento,_this.activa);
}

@override
String toString() {
  final _this = this as Tiquetera;
  return 'Tiquetera(id: ${_this.id}, almuerzosComprados: ${_this.almuerzosComprados}, saldo: ${_this.saldo}, fechaVencimiento: ${_this.fechaVencimiento}, activa: ${_this.activa})';
}


}

/// @nodoc
abstract mixin class $TiqueteraCopyWith<$Res>  {
  factory $TiqueteraCopyWith(Tiquetera value, $Res Function(Tiquetera) _then) = _$TiqueteraCopyWithImpl;
@useResult
$Res call({
 String id,@JsonKey(name: 'almuerzos_comprados') int almuerzosComprados, int saldo,@JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate) DateTime fechaVencimiento, bool activa
});




}
/// @nodoc
class _$TiqueteraCopyWithImpl<$Res>
    implements $TiqueteraCopyWith<$Res> {
  _$TiqueteraCopyWithImpl(this._self, this._then);

  final Tiquetera _self;
  final $Res Function(Tiquetera) _then;

/// Create a copy of Tiquetera
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? almuerzosComprados = null,Object? saldo = null,Object? fechaVencimiento = null,Object? activa = null,}) {
  return _then(Tiquetera(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,almuerzosComprados: null == almuerzosComprados ? _self.almuerzosComprados : almuerzosComprados // ignore: cast_nullable_to_non_nullable
as int,saldo: null == saldo ? _self.saldo : saldo // ignore: cast_nullable_to_non_nullable
as int,fechaVencimiento: null == fechaVencimiento ? _self.fechaVencimiento : fechaVencimiento // ignore: cast_nullable_to_non_nullable
as DateTime,activa: null == activa ? _self.activa : activa // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [Tiquetera].
extension TiqueteraPatterns on Tiquetera {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Tiquetera value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Tiquetera() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Tiquetera value)  $default,){
final _that = this;
switch (_that) {
case _Tiquetera():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Tiquetera value)?  $default,){
final _that = this;
switch (_that) {
case _Tiquetera() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id, @JsonKey(name: 'almuerzos_comprados')  int almuerzosComprados,  int saldo, @JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate)  DateTime fechaVencimiento,  bool activa)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Tiquetera() when $default != null:
return $default(_that.id,_that.almuerzosComprados,_that.saldo,_that.fechaVencimiento,_that.activa);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id, @JsonKey(name: 'almuerzos_comprados')  int almuerzosComprados,  int saldo, @JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate)  DateTime fechaVencimiento,  bool activa)  $default,) {final _that = this;
switch (_that) {
case _Tiquetera():
return $default(_that.id,_that.almuerzosComprados,_that.saldo,_that.fechaVencimiento,_that.activa);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id, @JsonKey(name: 'almuerzos_comprados')  int almuerzosComprados,  int saldo, @JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate)  DateTime fechaVencimiento,  bool activa)?  $default,) {final _that = this;
switch (_that) {
case _Tiquetera() when $default != null:
return $default(_that.id,_that.almuerzosComprados,_that.saldo,_that.fechaVencimiento,_that.activa);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Tiquetera implements Tiquetera {
  const _Tiquetera({required this.id, @JsonKey(name: 'almuerzos_comprados') required this.almuerzosComprados, required this.saldo, @JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate) required this.fechaVencimiento, required this.activa});
  factory _Tiquetera.fromJson(Map<String, dynamic> json) => _$TiqueteraFromJson(json);

@override final  String id;
@override@JsonKey(name: 'almuerzos_comprados') final  int almuerzosComprados;
@override final  int saldo;
@override@JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate) final  DateTime fechaVencimiento;
@override final  bool activa;

/// Create a copy of Tiquetera
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TiqueteraCopyWith<_Tiquetera> get copyWith => __$TiqueteraCopyWithImpl<_Tiquetera>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TiqueteraToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Tiquetera&&(identical(other.id, id) || other.id == id)&&(identical(other.almuerzosComprados, almuerzosComprados) || other.almuerzosComprados == almuerzosComprados)&&(identical(other.saldo, saldo) || other.saldo == saldo)&&(identical(other.fechaVencimiento, fechaVencimiento) || other.fechaVencimiento == fechaVencimiento)&&(identical(other.activa, activa) || other.activa == activa));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,almuerzosComprados,saldo,fechaVencimiento,activa);
}

@override
String toString() {
    return 'Tiquetera(id: $id, almuerzosComprados: $almuerzosComprados, saldo: $saldo, fechaVencimiento: $fechaVencimiento, activa: $activa)';
}


}

/// @nodoc
abstract mixin class _$TiqueteraCopyWith<$Res> implements $TiqueteraCopyWith<$Res> {
  factory _$TiqueteraCopyWith(_Tiquetera value, $Res Function(_Tiquetera) _then) = __$TiqueteraCopyWithImpl;
@override @useResult
$Res call({
 String id,@JsonKey(name: 'almuerzos_comprados') int almuerzosComprados, int saldo,@JsonKey(name: 'fecha_vencimiento', fromJson: parseApiDate) DateTime fechaVencimiento, bool activa
});




}
/// @nodoc
class __$TiqueteraCopyWithImpl<$Res>
    implements _$TiqueteraCopyWith<$Res> {
  __$TiqueteraCopyWithImpl(this._self, this._then);

  final _Tiquetera _self;
  final $Res Function(_Tiquetera) _then;

/// Create a copy of Tiquetera
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? almuerzosComprados = null,Object? saldo = null,Object? fechaVencimiento = null,Object? activa = null,}) {
  return _then(_Tiquetera(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,almuerzosComprados: null == almuerzosComprados ? _self.almuerzosComprados : almuerzosComprados // ignore: cast_nullable_to_non_nullable
as int,saldo: null == saldo ? _self.saldo : saldo // ignore: cast_nullable_to_non_nullable
as int,fechaVencimiento: null == fechaVencimiento ? _self.fechaVencimiento : fechaVencimiento // ignore: cast_nullable_to_non_nullable
as DateTime,activa: null == activa ? _self.activa : activa // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}


/// @nodoc
mixin _$Cliente {

 String get id; String get nombre; String get telefono; bool get activo; String? get documento; String? get email; Tiquetera? get tiquetera;
/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ClienteCopyWith<Cliente> get copyWith => _$ClienteCopyWithImpl<Cliente>(this as Cliente, _$identity);

  /// Serializes this Cliente to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Cliente;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Cliente&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.nombre, _this.nombre) || other.nombre == _this.nombre)&&(identical(other.telefono, _this.telefono) || other.telefono == _this.telefono)&&(identical(other.activo, _this.activo) || other.activo == _this.activo)&&(identical(other.documento, _this.documento) || other.documento == _this.documento)&&(identical(other.email, _this.email) || other.email == _this.email)&&(identical(other.tiquetera, _this.tiquetera) || other.tiquetera == _this.tiquetera));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Cliente;
  return Object.hash(runtimeType,_this.id,_this.nombre,_this.telefono,_this.activo,_this.documento,_this.email,_this.tiquetera);
}

@override
String toString() {
  final _this = this as Cliente;
  return 'Cliente(id: ${_this.id}, nombre: ${_this.nombre}, telefono: ${_this.telefono}, activo: ${_this.activo}, documento: ${_this.documento}, email: ${_this.email}, tiquetera: ${_this.tiquetera})';
}


}

/// @nodoc
abstract mixin class $ClienteCopyWith<$Res>  {
  factory $ClienteCopyWith(Cliente value, $Res Function(Cliente) _then) = _$ClienteCopyWithImpl;
@useResult
$Res call({
 String id, String nombre, String telefono, bool activo, String? documento, String? email, Tiquetera? tiquetera
});


$TiqueteraCopyWith<$Res>? get tiquetera;

}
/// @nodoc
class _$ClienteCopyWithImpl<$Res>
    implements $ClienteCopyWith<$Res> {
  _$ClienteCopyWithImpl(this._self, this._then);

  final Cliente _self;
  final $Res Function(Cliente) _then;

/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? nombre = null,Object? telefono = null,Object? activo = null,Object? documento = freezed,Object? email = freezed,Object? tiquetera = freezed,}) {
  return _then(Cliente(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,nombre: null == nombre ? _self.nombre : nombre // ignore: cast_nullable_to_non_nullable
as String,telefono: null == telefono ? _self.telefono : telefono // ignore: cast_nullable_to_non_nullable
as String,activo: null == activo ? _self.activo : activo // ignore: cast_nullable_to_non_nullable
as bool,documento: freezed == documento ? _self.documento : documento // ignore: cast_nullable_to_non_nullable
as String?,email: freezed == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String?,tiquetera: freezed == tiquetera ? _self.tiquetera : tiquetera // ignore: cast_nullable_to_non_nullable
as Tiquetera?,
  ));
}
/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TiqueteraCopyWith<$Res>? get tiquetera {
    if (_self.tiquetera == null) {
    return null;
  }

  return $TiqueteraCopyWith<$Res>(_self.tiquetera!, (value) {
    return _then(_self.copyWith(tiquetera: value));
  });
}
}


/// Adds pattern-matching-related methods to [Cliente].
extension ClientePatterns on Cliente {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Cliente value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Cliente() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Cliente value)  $default,){
final _that = this;
switch (_that) {
case _Cliente():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Cliente value)?  $default,){
final _that = this;
switch (_that) {
case _Cliente() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  String nombre,  String telefono,  bool activo,  String? documento,  String? email,  Tiquetera? tiquetera)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Cliente() when $default != null:
return $default(_that.id,_that.nombre,_that.telefono,_that.activo,_that.documento,_that.email,_that.tiquetera);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  String nombre,  String telefono,  bool activo,  String? documento,  String? email,  Tiquetera? tiquetera)  $default,) {final _that = this;
switch (_that) {
case _Cliente():
return $default(_that.id,_that.nombre,_that.telefono,_that.activo,_that.documento,_that.email,_that.tiquetera);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  String nombre,  String telefono,  bool activo,  String? documento,  String? email,  Tiquetera? tiquetera)?  $default,) {final _that = this;
switch (_that) {
case _Cliente() when $default != null:
return $default(_that.id,_that.nombre,_that.telefono,_that.activo,_that.documento,_that.email,_that.tiquetera);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Cliente implements Cliente {
  const _Cliente({required this.id, required this.nombre, required this.telefono, required this.activo, this.documento, this.email, this.tiquetera});
  factory _Cliente.fromJson(Map<String, dynamic> json) => _$ClienteFromJson(json);

@override final  String id;
@override final  String nombre;
@override final  String telefono;
@override final  bool activo;
@override final  String? documento;
@override final  String? email;
@override final  Tiquetera? tiquetera;

/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ClienteCopyWith<_Cliente> get copyWith => __$ClienteCopyWithImpl<_Cliente>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ClienteToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Cliente&&(identical(other.id, id) || other.id == id)&&(identical(other.nombre, nombre) || other.nombre == nombre)&&(identical(other.telefono, telefono) || other.telefono == telefono)&&(identical(other.activo, activo) || other.activo == activo)&&(identical(other.documento, documento) || other.documento == documento)&&(identical(other.email, email) || other.email == email)&&(identical(other.tiquetera, tiquetera) || other.tiquetera == tiquetera));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,nombre,telefono,activo,documento,email,tiquetera);
}

@override
String toString() {
    return 'Cliente(id: $id, nombre: $nombre, telefono: $telefono, activo: $activo, documento: $documento, email: $email, tiquetera: $tiquetera)';
}


}

/// @nodoc
abstract mixin class _$ClienteCopyWith<$Res> implements $ClienteCopyWith<$Res> {
  factory _$ClienteCopyWith(_Cliente value, $Res Function(_Cliente) _then) = __$ClienteCopyWithImpl;
@override @useResult
$Res call({
 String id, String nombre, String telefono, bool activo, String? documento, String? email, Tiquetera? tiquetera
});


@override $TiqueteraCopyWith<$Res>? get tiquetera;

}
/// @nodoc
class __$ClienteCopyWithImpl<$Res>
    implements _$ClienteCopyWith<$Res> {
  __$ClienteCopyWithImpl(this._self, this._then);

  final _Cliente _self;
  final $Res Function(_Cliente) _then;

/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? nombre = null,Object? telefono = null,Object? activo = null,Object? documento = freezed,Object? email = freezed,Object? tiquetera = freezed,}) {
  return _then(_Cliente(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,nombre: null == nombre ? _self.nombre : nombre // ignore: cast_nullable_to_non_nullable
as String,telefono: null == telefono ? _self.telefono : telefono // ignore: cast_nullable_to_non_nullable
as String,activo: null == activo ? _self.activo : activo // ignore: cast_nullable_to_non_nullable
as bool,documento: freezed == documento ? _self.documento : documento // ignore: cast_nullable_to_non_nullable
as String?,email: freezed == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String?,tiquetera: freezed == tiquetera ? _self.tiquetera : tiquetera // ignore: cast_nullable_to_non_nullable
as Tiquetera?,
  ));
}

/// Create a copy of Cliente
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TiqueteraCopyWith<$Res>? get tiquetera {
    if (_self.tiquetera == null) {
    return null;
  }

  return $TiqueteraCopyWith<$Res>(_self.tiquetera!, (value) {
    return _then(_self.copyWith(tiquetera: value));
  });
}
}

// dart format on
