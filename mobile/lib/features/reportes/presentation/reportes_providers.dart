import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/reportes/data/reportes_repository.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';

part 'reportes_providers.g.dart';

@Riverpod(keepAlive: true)
ReportesRepository reportesRepository(Ref ref) =>
    ReportesRepository(ref.watch(apiClientProvider));

@Riverpod(retry: noRetry)
Future<Dashboard> dashboard(Ref ref) =>
    ref.watch(reportesRepositoryProvider).dashboard();

class ReporteListState {
  const new({
    required this.items,
    required this.total,
    required this.page,
    this.loadingMore = false,
  });

  final List<ClienteReporte> items;
  final int total;
  final int page;
  final bool loadingMore;

  bool get hasMore => items.length < total;

  ReporteListState copyWith({required bool loadingMore}) => ReporteListState(
    items: items,
    total: total,
    page: page,
    loadingMore: loadingMore,
  );
}

/// Listado paginado de un reporte de clientes con su filtro [valor].
@Riverpod(retry: noRetry)
class ReporteList extends _$ReporteList {
  @override
  Future<ReporteListState> build(ReporteTipo tipo, int valor) async {
    final p = await ref.read(reportesRepositoryProvider).list(tipo, valor);
    return ReporteListState(items: p.items, total: p.total, page: p.page);
  }

  Future<void> loadMore() async {
    final c = state.value;
    if (c == null || c.loadingMore || !c.hasMore) return;
    state = AsyncData(c.copyWith(loadingMore: true));
    try {
      final next = await ref
          .read(reportesRepositoryProvider)
          .list(tipo, valor, page: c.page + 1);
      state = AsyncData(
        ReporteListState(
          items: [...c.items, ...next.items],
          total: next.total,
          page: next.page,
        ),
      );
    } on Object {
      state = AsyncData(c);
    }
  }
}

@Riverpod(retry: noRetry)
Future<ConsumosReporte> consumosReporte(Ref ref, RangoFechas rango) =>
    ref.watch(reportesRepositoryProvider).consumos(rango);
