sealed class AppFailure implements Exception {
  const new();
}

final class NetworkFailure extends AppFailure {
  const new();
}

final class TimeoutFailure extends AppFailure {
  const new();
}

final class UnauthorizedFailure extends AppFailure {
  const new({this.code});
  final String? code;
}

final class ApiFailure extends AppFailure {
  const new({required this.status, required this.code, this.detail});
  final int status;
  final String code;
  final String? detail;
}

final class UnknownFailure extends AppFailure {
  const new();
}
