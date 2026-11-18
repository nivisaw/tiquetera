import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/clientes/presentation/widgets.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_providers.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

/// Dashboard del admin y acceso a los reportes.
class ReportesPage extends ConsumerWidget {
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.reportesTitle)),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(dashboardProvider.future),
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.md),
          children: [
            switch (ref.watch(dashboardProvider)) {
              AsyncData(:final value) => _DashboardView(d: value),
              AsyncError(:final error) => ErrorView(
                error: error,
                onRetry: () => ref.invalidate(dashboardProvider),
              ),
              _ => Padding(
                padding: const EdgeInsets.all(AppSpacing.lg),
                child: Center(
                  child: CircularProgressIndicator(
                    semanticsLabel: l10n.loading,
                  ),
                ),
              ),
            },
            const SizedBox(height: AppSpacing.lg),
            const _ReportLinks(),
          ],
        ),
      ),
    );
  }
}

class _DashboardView extends StatelessWidget {
  const new({required this.d});

  final Dashboard d;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final stats = [
      (Icons.restaurant, l10n.dashConsumosHoy, d.consumosHoy),
      (
        Icons.people_alt_outlined,
        l10n.dashClientesVigentes,
        d.clientesVigentes,
      ),
      (Icons.inventory_2_outlined, l10n.dashSaldoPendiente, d.saldoPendiente),
      (Icons.add_card, l10n.dashVendidosMes, d.vendidosMes),
      (Icons.event_busy, l10n.dashPorVencer, d.porVencer7d),
      (Icons.block, l10n.dashRechazosHoy, d.rechazosHoy),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Wrap(
          spacing: AppSpacing.sm,
          runSpacing: AppSpacing.sm,
          children: [
            for (final (icon, label, value) in stats)
              _StatCard(icon: icon, label: label, value: value),
          ],
        ),
        const SizedBox(height: AppSpacing.lg),
        Semantics(
          header: true,
          child: Text(
            l10n.dashUltimos7,
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        DayBars(days: d.consumos7d),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  const new({required this.icon, required this.label, required this.value});

  final IconData icon;
  final String label;
  final int value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final width =
        (MediaQuery.sizeOf(context).width - AppSpacing.md * 2 - AppSpacing.sm) /
        2;
    return SizedBox(
      width: width,
      child: Card(
        margin: EdgeInsets.zero,
        child: Semantics(
          container: true,
          label: '$label: $value',
          excludeSemantics: true,
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                DecoratedBox(
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(AppRadius.md),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(AppSpacing.sm),
                    child: Icon(
                      icon,
                      color: theme.colorScheme.onPrimaryContainer,
                    ),
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                Text('$value', style: theme.textTheme.headlineMedium),
                Text(label, style: theme.textTheme.bodyMedium),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Barras horizontales por día: legibles con texto grande y sin librería
/// de gráficos.
class DayBars extends StatelessWidget {
  const new({required this.days, super.key});

  final List<DiaTotal> days;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final maxTotal = days.fold<int>(0, (m, d) => d.total > m ? d.total : m);
    return Column(
      children: [
        for (final d in days)
          Semantics(
            container: true,
            label: l10n.dashBarra(ClienteRules.displayDate(d.fecha), d.total),
            excludeSemantics: true,
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
              child: Row(
                children: [
                  SizedBox(
                    width: 64,
                    child: Text(
                      ClienteRules.displayDate(d.fecha).substring(0, 5),
                      style: theme.textTheme.bodySmall,
                    ),
                  ),
                  Expanded(
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        value: maxTotal == 0 ? 0 : d.total / maxTotal,
                        minHeight: 16,
                      ),
                    ),
                  ),
                  SizedBox(
                    width: 48,
                    child: Text(
                      '${d.total}',
                      textAlign: TextAlign.end,
                      style: theme.textTheme.bodyMedium,
                    ),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

class _ReportLinks extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final links = [
      (
        Icons.history_toggle_off,
        l10n.reporteSinRecargaTitle,
        l10n.reporteSinRecargaDesc,
        '/reportes/${ReporteTipo.sinRecarga.path}',
      ),
      (
        Icons.event_busy,
        l10n.reportePorVencerTitle,
        l10n.reportePorVencerDesc,
        '/reportes/${ReporteTipo.porVencer.path}',
      ),
      (
        Icons.battery_alert,
        l10n.reporteSaldoBajoTitle,
        l10n.reporteSaldoBajoDesc,
        '/reportes/${ReporteTipo.saldoBajo.path}',
      ),
      (
        Icons.point_of_sale,
        l10n.reporteConsumosTitle,
        l10n.reporteConsumosDesc,
        '/reportes/consumos',
      ),
    ];
    return Card(
      margin: EdgeInsets.zero,
      child: Column(
        children: [
          for (final (icon, title, desc, path) in links)
            ListTile(
              key: Key('link_$path'),
              leading: Icon(icon),
              title: Text(title),
              subtitle: Text(desc),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push(path),
            ),
        ],
      ),
    );
  }
}
