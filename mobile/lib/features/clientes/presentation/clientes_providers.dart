import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/features/clientes/data/clientes_repository.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';

part 'clientes_providers.g.dart';

@Riverpod(keepAlive: true)
ClientesRepository clientesRepository(Ref ref) =>
    ClientesRepository(ref.watch(apiClientProvider));

/// Sin reintento automático: la UI muestra el error con botón Reintentar.
Duration? noRetry(int retryCount, Object error) => null;

class ClientesListState {
  const new({
    required this.items,
    required this.total,
    required this.page,
    this.query = '',
    this.loadingMore = false,
  });

  final List<Cliente> items;
  final int total;
  final int page;
  final String query;
  final bool loadingMore;

  bool get hasMore => items.length < total;
}

/// Listado paginado con búsqueda; `search` reinicia desde la página 1.
@Riverpod(retry: noRetry)
class ClientesList extends _$ClientesList {
  String _query = '';

  @override
  Future<ClientesListState> build() => _fetchFirst();

  Future<ClientesListState> _fetchFirst() async {
    final q = _query;
    final page = await ref.read(clientesRepositoryProvider).list(query: q);
    return ClientesListState(
      items: page.items,
      total: page.total,
      page: page.page,
      query: q,
    );
  }

  Future<void> search(String query) async {
    if (query.trim() == _query && state.hasValue) return;
    _query = query.trim();
    state = const AsyncLoading();
    state = await AsyncValue.guard(_fetchFirst);
  }

  Future<void> retry() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(_fetchFirst);
  }

  Future<void> loadMore() async {
    final current = state.value;
    if (current == null || current.loadingMore || !current.hasMore) return;
    state = AsyncData(_copy(current, loadingMore: true));
    try {
      final next = await ref
          .read(clientesRepositoryProvider)
          .list(query: current.query, page: current.page + 1);
      state = AsyncData(
        ClientesListState(
          items: [...current.items, ...next.items],
          total: next.total,
          page: next.page,
          query: current.query,
        ),
      );
    } on Object {
      state = AsyncData(_copy(current, loadingMore: false));
    }
  }

  ClientesListState _copy(ClientesListState s, {required bool loadingMore}) =>
      ClientesListState(
        items: s.items,
        total: s.total,
        page: s.page,
        query: s.query,
        loadingMore: loadingMore,
      );
}

@Riverpod(retry: noRetry)
Future<Cliente> clienteDetalle(Ref ref, String id) =>
    ref.watch(clientesRepositoryProvider).get(id);
