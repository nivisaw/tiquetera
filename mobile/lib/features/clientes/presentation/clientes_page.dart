import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/auth/presentation/logout_button.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/clientes/presentation/widgets.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class ClientesPage extends ConsumerStatefulWidget {
  const new({super.key});

  static const Duration debounce = Duration(milliseconds: 400);

  @override
  ConsumerState<ClientesPage> createState() => _ClientesPageState();
}

class _ClientesPageState extends ConsumerState<ClientesPage> {
  Timer? _debounce;

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }

  void _onQuery(String q) {
    _debounce?.cancel();
    _debounce = Timer(
      ClientesPage.debounce,
      () => ref.read(clientesListProvider.notifier).search(q),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(clientesListProvider);
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.clientesTitle),
        actions: const [LogoutButton()],
      ),
      floatingActionButton: FloatingActionButton.extended(
        key: const Key('clientes_new'),
        onPressed: () => context.push('/clientes/nuevo'),
        icon: const Icon(Icons.person_add),
        label: Text(l10n.clientesNew),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: TextField(
              key: const Key('clientes_search'),
              onChanged: _onQuery,
              textInputAction: TextInputAction.search,
              decoration: InputDecoration(
                hintText: l10n.clientesSearch,
                prefixIcon: const Icon(Icons.search),
              ),
            ),
          ),
          Expanded(
            child: switch (list) {
              AsyncData(:final value) => _List(state: value),
              AsyncError(:final error) => ErrorView(
                error: error,
                onRetry: ref.read(clientesListProvider.notifier).retry,
              ),
              _ => Center(
                child: CircularProgressIndicator(semanticsLabel: l10n.loading),
              ),
            },
          ),
        ],
      ),
    );
  }
}

class _List extends ConsumerWidget {
  const new({required this.state});

  final ClientesListState state;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    if (state.items.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: Text(
            state.query.isEmpty ? l10n.clientesEmpty : l10n.clientesNoResults,
            textAlign: TextAlign.center,
          ),
        ),
      );
    }
    final extra = state.hasMore ? 1 : 0;
    return RefreshIndicator(
      onRefresh: ref.read(clientesListProvider.notifier).retry,
      child: ListView.builder(
        padding: const EdgeInsets.only(bottom: 88),
        itemCount: state.items.length + extra,
        itemBuilder: (context, i) {
          if (i < state.items.length) return _Tile(cliente: state.items[i]);
          return Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: state.loadingMore
                ? Center(
                    child: CircularProgressIndicator(
                      semanticsLabel: l10n.loading,
                    ),
                  )
                : OutlinedButton(
                    key: const Key('clientes_more'),
                    onPressed: ref.read(clientesListProvider.notifier).loadMore,
                    child: Text(l10n.clientesLoadMore),
                  ),
          );
        },
      ),
    );
  }
}

class _Tile extends StatelessWidget {
  const new({required this.cliente});

  final Cliente cliente;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListTile(
      minTileHeight: 56,
      leading: const Icon(Icons.person),
      title: Text(cliente.nombre),
      subtitle: Text(ClienteRules.displayTelefono(cliente.telefono)),
      trailing: cliente.activo
          ? const Icon(Icons.chevron_right)
          : Chip(label: Text(l10n.clienteInactivo)),
      onTap: () => context.push('/clientes/${cliente.id}'),
    );
  }
}
