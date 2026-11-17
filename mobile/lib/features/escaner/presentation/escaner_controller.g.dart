// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'escaner_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(consumosRepository)
final consumosRepositoryProvider = ConsumosRepositoryProvider._();

final class ConsumosRepositoryProvider
    extends
        $FunctionalProvider<
          ConsumosRepository,
          ConsumosRepository,
          ConsumosRepository
        >
    with $Provider<ConsumosRepository> {
  ConsumosRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'consumosRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$consumosRepositoryHash();

  @$internal
  @override
  $ProviderElement<ConsumosRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  ConsumosRepository create(Ref ref) {
    return consumosRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(ConsumosRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<ConsumosRepository>(value),
    );
  }
}

String _$consumosRepositoryHash() =>
    r'bea5952305234cf3509efa3675ae99593c2055a4';

@ProviderFor(cameraPermission)
final cameraPermissionProvider = CameraPermissionProvider._();

final class CameraPermissionProvider
    extends
        $FunctionalProvider<
          CameraPermission,
          CameraPermission,
          CameraPermission
        >
    with $Provider<CameraPermission> {
  CameraPermissionProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'cameraPermissionProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$cameraPermissionHash();

  @$internal
  @override
  $ProviderElement<CameraPermission> $createElement($ProviderPointer pointer) =>
      $ProviderElement(pointer);

  @override
  CameraPermission create(Ref ref) {
    return cameraPermission(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(CameraPermission value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<CameraPermission>(value),
    );
  }
}

String _$cameraPermissionHash() => r'b0c58ef5a25bb7f65e42563c2017314d9156b1df';

@ProviderFor(idempotencyKeyGenerator)
final idempotencyKeyGeneratorProvider = IdempotencyKeyGeneratorProvider._();

final class IdempotencyKeyGeneratorProvider
    extends
        $FunctionalProvider<
          String Function(),
          String Function(),
          String Function()
        >
    with $Provider<String Function()> {
  IdempotencyKeyGeneratorProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'idempotencyKeyGeneratorProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$idempotencyKeyGeneratorHash();

  @$internal
  @override
  $ProviderElement<String Function()> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  String Function() create(Ref ref) {
    return idempotencyKeyGenerator(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(String Function() value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<String Function()>(value),
    );
  }
}

String _$idempotencyKeyGeneratorHash() =>
    r'679f86b52c0536a23d8966c396b09e3e74cd3dd0';

@ProviderFor(EscanerController)
final escanerControllerProvider = EscanerControllerProvider._();

final class EscanerControllerProvider
    extends $NotifierProvider<EscanerController, EscanerState> {
  EscanerControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'escanerControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$escanerControllerHash();

  @$internal
  @override
  EscanerController create() => EscanerController();

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(EscanerState value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<EscanerState>(value),
    );
  }
}

String _$escanerControllerHash() => r'5e5094c790278c19d66c3453fb2c15547131bc46';

abstract class _$EscanerController extends $Notifier<EscanerState> {
  EscanerState build();
  @$mustCallSuper
  @override
  WhenComplete runBuild() {
    final ref = this.ref as $Ref<EscanerState, EscanerState>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<EscanerState, EscanerState>,
              EscanerState,
              Object?,
              Object?
            >;
    return element.handleCreate(ref, build);
  }
}
