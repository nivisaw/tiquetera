import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:tiquetera/core/storage/token_storage.dart';

typedef Responder = ({int status, Object? body}) Function(RequestOptions o);

class FakeAdapter implements HttpClientAdapter {
  new(this.responder);

  final Responder responder;
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    await Future<void>.delayed(Duration.zero);
    final r = responder(options);
    return ResponseBody.fromString(
      r.body == null ? '' : jsonEncode(r.body),
      r.status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

class MemoryTokenStorage implements TokenStorage {
  new([this.tokens]);

  Tokens? tokens;

  @override
  Future<Tokens?> read() async => tokens;

  @override
  Future<void> write(Tokens tokens) async => this.tokens = tokens;

  @override
  Future<void> clear() async => tokens = null;
}
