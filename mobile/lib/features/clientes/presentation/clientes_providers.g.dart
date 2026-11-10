// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'clientes_providers.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(clientesRepository)
final clientesRepositoryProvider = ClientesRepositoryProvider._();

final class ClientesRepositoryProvider
    extends
        $FunctionalProvider<
          ClientesRepository,
          ClientesRepository,
          ClientesRepository
        >
    with $Provider<ClientesRepository> {
  ClientesRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'clientesRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$clientesRepositoryHash();

  @$internal
  @override
  $ProviderElement<ClientesRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  ClientesRepository create(Ref ref) {
    return clientesRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(ClientesRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<ClientesRepository>(value),
    );
  }
}

String _$clientesRepositoryHash() =>
    r'929f962258cc4775b16780431ca8dd40d73515d3';

/// Listado paginado con búsqueda; `search` reinicia desde la página 1.

@ProviderFor(ClientesList)
final clientesListProvider = ClientesListProvider._();

/// Listado paginado con búsqueda; `search` reinicia desde la página 1.
final class ClientesListProvider
    extends $AsyncNotifierProvider<ClientesList, ClientesListState> {
  /// Listado paginado con búsqueda; `search` reinicia desde la página 1.
  ClientesListProvider._()
    : super(
        from: null,
        argument: null,
        retry: noRetry,
        name: r'clientesListProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$clientesListHash();

  @$internal
  @override
  ClientesList create() => ClientesList();
}

String _$clientesListHash() => r'931f6df64133f9b7ebed9a1cf0d37b6bde109a2c';

/// Listado paginado con búsqueda; `search` reinicia desde la página 1.

abstract class _$ClientesList extends $AsyncNotifier<ClientesListState> {
  FutureOr<ClientesListState> build();
  @$mustCallSuper
  @override
  WhenComplete runBuild() {
    final ref =
        this.ref as $Ref<AsyncValue<ClientesListState>, ClientesListState>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ClientesListState>, ClientesListState>,
              AsyncValue<ClientesListState>,
              Object?,
              Object?
            >;
    return element.handleCreate(ref, build);
  }
}

@ProviderFor(clienteDetalle)
final clienteDetalleProvider = ClienteDetalleFamily._();

final class ClienteDetalleProvider
    extends $FunctionalProvider<AsyncValue<Cliente>, Cliente, FutureOr<Cliente>>
    with $FutureModifier<Cliente>, $FutureProvider<Cliente> {
  ClienteDetalleProvider._({
    required ClienteDetalleFamily super.from,
    required String super.argument,
  }) : super(
         retry: noRetry,
         name: r'clienteDetalleProvider',
         isAutoDispose: true,
         dependencies: null,
         $allTransitiveDependencies: null,
       );

  @override
  String debugGetCreateSourceHash() => _$clienteDetalleHash();

  @override
  String toString() {
    return r'clienteDetalleProvider'
        ''
        '($argument)';
  }

  @$internal
  @override
  $FutureProviderElement<Cliente> $createElement($ProviderPointer pointer) =>
      $FutureProviderElement(pointer);

  @override
  FutureOr<Cliente> create(Ref ref) {
    final argument = this.argument as String;
    return clienteDetalle(ref, argument);
  }

  @override
  bool operator ==(Object other) {
    return other is ClienteDetalleProvider && other.argument == argument;
  }

  @override
  int get hashCode {
    return argument.hashCode;
  }
}

String _$clienteDetalleHash() => r'df1d80df1d252fb71907f4ff4439a0fbc77abf20';

final class ClienteDetalleFamily extends $Family
    with $FunctionalFamilyOverride<FutureOr<Cliente>, String> {
  ClienteDetalleFamily._()
    : super(
        retry: noRetry,
        name: r'clienteDetalleProvider',
        dependencies: null,
        $allTransitiveDependencies: null,
        isAutoDispose: true,
      );

  ClienteDetalleProvider call(String id) =>
      ClienteDetalleProvider._(argument: id, from: this);

  @override
  String toString() => r'clienteDetalleProvider';
}
