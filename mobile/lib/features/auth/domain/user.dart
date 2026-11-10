import 'package:freezed_annotation/freezed_annotation.dart';

part 'user.freezed.dart';
part 'user.g.dart';

enum Role { admin, cajero }

@freezed
abstract class User with _$User {
  const factory({
    required String id,
    required String nombre,
    @JsonKey(unknownEnumValue: JsonKey.nullForUndefinedEnumValue) Role? rol,
  }) = _User;

  factory fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
}
