import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/clientes/domain/whatsapp_link.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/clientes/presentation/widgets.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_providers.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

/// Listado de clientes de un reporte, con filtro y acceso al chat.
class ReporteClientesPage extends ConsumerStatefulWidget {
  const new({required this.tipo, super.key});

  final ReporteTipo tipo;

  @override
  ConsumerState<ReporteClientesPage> createState() => _State();
}

class _State extends ConsumerState<ReporteClientesPage> {
  late int _valor = widget.tipo.defaultValue;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final tipo = widget.tipo;
    final provider = reporteListProvider(tipo, _valor);
    return Scaffold(
      appBar: AppBar(title: Text(_title(l10n, tipo))),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.md,
              AppSpacing.sm,
              AppSpacing.md,
              0,
            ),
            child: Text(_filterLabel(l10n, tipo)),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
            child: Wrap(
              spacing: AppSpacing.sm,
              children: [
                for (final v in tipo.opciones)
                  ChoiceChip(
                    key: Key('filtro_$v'),
                    label: Text(
                      tipo == ReporteTipo.saldoBajo
                          ? l10n.filtroAlmuerzos(v)
                          : l10n.filtroDias(v),
                    ),
                    selected: v == _valor,
                    onSelected: (_) => setState(() => _valor = v),
                  ),
              ],
            ),
          ),
          const Divider(),
          Expanded(
            child: switch (ref.watch(provider)) {
              AsyncData(:final value) when value.items.isEmpty => Center(
                child: Padding(
                  padding: const EdgeInsets.all(AppSpacing.lg),
                  child: Text(l10n.reporteEmpty, textAlign: TextAlign.center),
                ),
              ),
              AsyncData(:final value) => _List(
                tipo: tipo,
                state: value,
                onMore: () => ref.read(provider.notifier).loadMore(),
              ),
              AsyncError(:final error) => ErrorView(
                error: error,
                onRetry: () => ref.invalidate(provider),
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
  const new({required this.tipo, required this.state, required this.onMore});

  final ReporteTipo tipo;
  final ReporteListState state;
  final VoidCallback onMore;

  Future<void> _chat(
    BuildContext context,
    WidgetRef ref,
    ClienteReporte c,
  ) async {
    final l10n = AppLocalizations.of(context);
    final uri = WhatsAppLink.build(c.telefono, _message(l10n, tipo, c));
    final ok =
        uri != null && await ref.read(qrShareServiceProvider).openChat(uri);
    if (!ok && context.mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.qrChatError)));
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final items = state.items;
    return ListView.builder(
      itemCount: items.length + 2,
      itemBuilder: (context, i) {
        if (i == 0) {
          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
            child: Text(
              l10n.reporteTotal(state.total),
              style: Theme.of(context).textTheme.labelLarge,
            ),
          );
        }
        if (i == items.length + 1) {
          return state.hasMore
              ? Padding(
                  padding: const EdgeInsets.all(AppSpacing.sm),
                  child: TextButton(
                    key: const Key('reporte_more'),
                    onPressed: state.loadingMore ? null : onMore,
                    child: Text(l10n.clientesLoadMore),
                  ),
                )
              : const SizedBox(height: AppSpacing.lg);
        }
        final c = items[i - 1];
        final fecha = ClienteRules.displayDate(c.fecha);
        final detalle = tipo == ReporteTipo.sinRecarga
            ? l10n.itemUltimaRecarga(fecha)
            : l10n.itemVence(fecha);
        return ListTile(
          key: ValueKey('reporte_${c.clienteId}'),
          title: Text(c.nombre),
          subtitle: Text(
            '${ClienteRules.displayTelefono(c.telefono)}\n'
            '$detalle · ${l10n.saldoLabel(c.saldo)}',
          ),
          isThreeLine: true,
          onTap: () => unawaited(context.push('/clientes/${c.clienteId}')),
          trailing: IconButton(
            key: Key('chat_${c.clienteId}'),
            tooltip: l10n.qrOpenChat,
            icon: const Icon(Icons.chat_outlined),
            onPressed: () => _chat(context, ref, c),
          ),
        );
      },
    );
  }
}

String _title(AppLocalizations l10n, ReporteTipo t) => switch (t) {
  ReporteTipo.sinRecarga => l10n.reporteSinRecargaTitle,
  ReporteTipo.porVencer => l10n.reportePorVencerTitle,
  ReporteTipo.saldoBajo => l10n.reporteSaldoBajoTitle,
};

String _filterLabel(AppLocalizations l10n, ReporteTipo t) => switch (t) {
  ReporteTipo.sinRecarga => l10n.filtroSinRecarga,
  ReporteTipo.porVencer => l10n.filtroPorVencer,
  ReporteTipo.saldoBajo => l10n.filtroSaldoBajo,
};

String _message(AppLocalizations l10n, ReporteTipo t, ClienteReporte c) {
  final nombre = c.nombre.split(' ').first;
  return switch (t) {
    ReporteTipo.sinRecarga => l10n.msgSinRecarga(nombre),
    ReporteTipo.porVencer => l10n.msgPorVencer(
      nombre,
      ClienteRules.displayDate(c.fecha),
      c.saldo,
    ),
    ReporteTipo.saldoBajo => l10n.msgSaldoBajo(nombre, c.saldo),
  };
}
