import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/features/escaner/data/camera_permission.dart';
import 'package:tiquetera/features/escaner/data/consumos_repository.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';
import 'package:tiquetera/features/escaner/domain/idempotency_key.dart';
import 'package:tiquetera/features/escaner/domain/scan_debouncer.dart';

part 'escaner_controller.g.dart';

@Riverpod(keepAlive: true)
ConsumosRepository consumosRepository(Ref ref) =>
    ConsumosRepository(ref.watch(apiClientProvider));

@Riverpod(keepAlive: true)
CameraPermission cameraPermission(Ref ref) => const CameraPermission();

@Riverpod(keepAlive: true)
String Function() idempotencyKeyGenerator(Ref ref) => newIdempotencyKey;

sealed class EscanerState {
  const new();
}

final class Escaneando extends EscanerState {
  const new();
}

final class Procesando extends EscanerState {
  const new();
}

final class Exito extends EscanerState {
  const new(this.consumo);
  final ConsumoOk consumo;
}

final class Rechazo extends EscanerState {
  const new(this.motivo);
  final RechazoMotivo motivo;
}

@riverpod
class EscanerController extends _$EscanerController {
  late ScanDebouncer _debouncer;

  @override
  EscanerState build() {
    _debouncer = ScanDebouncer(ref.watch(clockProvider));
    return const Escaneando();
  }

  Future<void> onScanned(String raw) async {
    final token = raw.trim();
    if (state is! Escaneando || token.isEmpty) return;
    if (!_debouncer.accept(token)) return;
    state = const Procesando();
    try {
      final ok = await ref
          .read(consumosRepositoryProvider)
          .registrar(
            qr: token,
            idempotencyKey: ref.read(idempotencyKeyGeneratorProvider)(),
          );
      if (ref.mounted) state = Exito(ok);
    } on AppFailure catch (f) {
      final motivo = _motivo(f);
      if (motivo == RechazoMotivo.sinRed) _debouncer.forget(token);
      if (ref.mounted) state = Rechazo(motivo);
    }
  }

  void volver() => state = const Escaneando();

  RechazoMotivo _motivo(AppFailure f) => switch (f) {
    NetworkFailure() || TimeoutFailure() => RechazoMotivo.sinRed,
    ApiFailure(status: 422, :final code) => RechazoMotivo.fromCode(code),
    ApiFailure(status: 429) => RechazoMotivo.demasiadosIntentos,
    _ => RechazoMotivo.error,
  };
}
