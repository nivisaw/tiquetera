// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'reportes_providers.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(reportesRepository)
final reportesRepositoryProvider = ReportesRepositoryProvider._();

final class ReportesRepositoryProvider
    extends
        $FunctionalProvider<
          ReportesRepository,
          ReportesRepository,
          ReportesRepository
        >
    with $Provider<ReportesRepository> {
  ReportesRepositoryProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'reportesRepositoryProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$reportesRepositoryHash();

  @$internal
  @override
  $ProviderElement<ReportesRepository> $createElement(
    $ProviderPointer pointer,
  ) => $ProviderElement(pointer);

  @override
  ReportesRepository create(Ref ref) {
    return reportesRepository(ref);
  }

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(ReportesRepository value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<ReportesRepository>(value),
    );
  }
}

String _$reportesRepositoryHash() =>
    r'7d6bcea3fe2e87dbd6ae2a80d3c38042e1c3b670';

@ProviderFor(dashboard)
final dashboardProvider = DashboardProvider._();

final class DashboardProvider
    extends
        $FunctionalProvider<
          AsyncValue<Dashboard>,
          Dashboard,
          FutureOr<Dashboard>
        >
    with $FutureModifier<Dashboard>, $FutureProvider<Dashboard> {
  DashboardProvider._()
    : super(
        from: null,
        argument: null,
        retry: noRetry,
        name: r'dashboardProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$dashboardHash();

  @$internal
  @override
  $FutureProviderElement<Dashboard> $createElement($ProviderPointer pointer) =>
      $FutureProviderElement(pointer);

  @override
  FutureOr<Dashboard> create(Ref ref) {
    return dashboard(ref);
  }
}

String _$dashboardHash() => r'fb46f9f0bfdbd6f44279295b548458437b1d4ba7';

/// Listado paginado de un reporte de clientes con su filtro [valor].

@ProviderFor(ReporteList)
final reporteListProvider = ReporteListFamily._();

/// Listado paginado de un reporte de clientes con su filtro [valor].
final class ReporteListProvider
    extends $AsyncNotifierProvider<ReporteList, ReporteListState> {
  /// Listado paginado de un reporte de clientes con su filtro [valor].
  ReporteListProvider._({
    required ReporteListFamily super.from,
    required (ReporteTipo, int) super.argument,
  }) : super(
         retry: noRetry,
         name: r'reporteListProvider',
         isAutoDispose: true,
         dependencies: null,
         $allTransitiveDependencies: null,
       );

  @override
  String debugGetCreateSourceHash() => _$reporteListHash();

  @override
  String toString() {
    return r'reporteListProvider'
        ''
        '$argument';
  }

  @$internal
  @override
  ReporteList create() => ReporteList();

  @override
  bool operator ==(Object other) {
    return other is ReporteListProvider && other.argument == argument;
  }

  @override
  int get hashCode {
    return argument.hashCode;
  }
}

String _$reporteListHash() => r'cde93b7f31cde54c0d3071e5229baf593ddb9479';

/// Listado paginado de un reporte de clientes con su filtro [valor].

final class ReporteListFamily extends $Family
    with
        $ClassFamilyOverride<
          ReporteList,
          AsyncValue<ReporteListState>,
          ReporteListState,
          FutureOr<ReporteListState>,
          (ReporteTipo, int)
        > {
  ReporteListFamily._()
    : super(
        retry: noRetry,
        name: r'reporteListProvider',
        dependencies: null,
        $allTransitiveDependencies: null,
        isAutoDispose: true,
      );

  /// Listado paginado de un reporte de clientes con su filtro [valor].

  ReporteListProvider call(ReporteTipo tipo, int valor) =>
      ReporteListProvider._(argument: (tipo, valor), from: this);

  @override
  String toString() => r'reporteListProvider';
}

/// Listado paginado de un reporte de clientes con su filtro [valor].

abstract class _$ReporteList extends $AsyncNotifier<ReporteListState> {
  late final _$args = ref.$arg as (ReporteTipo, int);
  ReporteTipo get tipo => _$args.$1;
  int get valor => _$args.$2;

  FutureOr<ReporteListState> build(ReporteTipo tipo, int valor);
  @$mustCallSuper
  @override
  WhenComplete runBuild() {
    final ref =
        this.ref as $Ref<AsyncValue<ReporteListState>, ReporteListState>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ReporteListState>, ReporteListState>,
              AsyncValue<ReporteListState>,
              Object?,
              Object?
            >;
    return element.handleCreate(ref, () => build(_$args.$1, _$args.$2));
  }
}

@ProviderFor(consumosReporte)
final consumosReporteProvider = ConsumosReporteFamily._();

final class ConsumosReporteProvider
    extends
        $FunctionalProvider<
          AsyncValue<ConsumosReporte>,
          ConsumosReporte,
          FutureOr<ConsumosReporte>
        >
    with $FutureModifier<ConsumosReporte>, $FutureProvider<ConsumosReporte> {
  ConsumosReporteProvider._({
    required ConsumosReporteFamily super.from,
    required RangoFechas super.argument,
  }) : super(
         retry: noRetry,
         name: r'consumosReporteProvider',
         isAutoDispose: true,
         dependencies: null,
         $allTransitiveDependencies: null,
       );

  @override
  String debugGetCreateSourceHash() => _$consumosReporteHash();

  @override
  String toString() {
    return r'consumosReporteProvider'
        ''
        '($argument)';
  }

  @$internal
  @override
  $FutureProviderElement<ConsumosReporte> $createElement(
    $ProviderPointer pointer,
  ) => $FutureProviderElement(pointer);

  @override
  FutureOr<ConsumosReporte> create(Ref ref) {
    final argument = this.argument as RangoFechas;
    return consumosReporte(ref, argument);
  }

  @override
  bool operator ==(Object other) {
    return other is ConsumosReporteProvider && other.argument == argument;
  }

  @override
  int get hashCode {
    return argument.hashCode;
  }
}

String _$consumosReporteHash() => r'9aa6b03180fc4cd53af1c017837c1a51e10311e3';

final class ConsumosReporteFamily extends $Family
    with $FunctionalFamilyOverride<FutureOr<ConsumosReporte>, RangoFechas> {
  ConsumosReporteFamily._()
    : super(
        retry: noRetry,
        name: r'consumosReporteProvider',
        dependencies: null,
        $allTransitiveDependencies: null,
        isAutoDispose: true,
      );

  ConsumosReporteProvider call(RangoFechas rango) =>
      ConsumosReporteProvider._(argument: rango, from: this);

  @override
  String toString() => r'consumosReporteProvider';
}
