import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/clientes/presentation/widgets.dart';
import 'package:tiquetera/features/reportes/domain/reportes.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_page.dart';
import 'package:tiquetera/features/reportes/presentation/reportes_providers.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

/// Consumos por día y por cajero en un rango (por defecto, última semana).
class ConsumosReportePage extends ConsumerStatefulWidget {
  const new({super.key});

  @override
  ConsumerState<ConsumosReportePage> createState() => _State();
}

class _State extends ConsumerState<ConsumosReportePage> {
  late final DateTime _hoy = ClienteRules.todayBogota(
    ref.read(clockProvider)(),
  );
  late RangoFechas _rango = RangoFechas.ultimaSemana(_hoy);

  Future<void> _pickRange() async {
    final l10n = AppLocalizations.of(context);
    final picked = await showDateRangePicker(
      context: context,
      firstDate: DateTime(_hoy.year - 2),
      lastDate: _hoy,
      initialDateRange: DateTimeRange(start: _rango.desde, end: _rango.hasta),
    );
    if (picked == null || !mounted) return;
    final rango = RangoFechas(picked.start, picked.end);
    if (!rango.valido) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.consumosRangoInvalido)));
      return;
    }
    setState(() => _rango = rango);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final provider = consumosReporteProvider(_rango);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.reporteConsumosTitle)),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.md),
        children: [
          Text(
            l10n.consumosRango(
              ClienteRules.displayDate(_rango.desde),
              ClienteRules.displayDate(_rango.hasta),
            ),
            style: Theme.of(context).textTheme.titleMedium,
          ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              key: const Key('consumos_rango'),
              onPressed: _pickRange,
              icon: const Icon(Icons.date_range),
              label: Text(l10n.consumosCambiarRango),
            ),
          ),
          switch (ref.watch(provider)) {
            AsyncData(:final value) => _Body(r: value),
            AsyncError(:final error) => ErrorView(
              error: error,
              onRetry: () => ref.invalidate(provider),
            ),
            _ => Padding(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Center(
                child: CircularProgressIndicator(semanticsLabel: l10n.loading),
              ),
            ),
          },
        ],
      ),
    );
  }
}

class _Body extends StatelessWidget {
  const new({required this.r});

  final ConsumosReporte r;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    if (r.total == 0) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.lg),
        child: Text(l10n.consumosSinDatos),
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(l10n.consumosTotal(r.total), style: text.headlineSmall),
        const SizedBox(height: AppSpacing.lg),
        Semantics(
          header: true,
          child: Text(l10n.consumosPorCajero, style: text.titleMedium),
        ),
        for (final c in r.cajeros)
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.person_outline),
            title: Text(c.nombre),
            trailing: Text('${c.total}', style: text.titleMedium),
          ),
        const SizedBox(height: AppSpacing.md),
        Semantics(
          header: true,
          child: Text(l10n.consumosPorDia, style: text.titleMedium),
        ),
        const SizedBox(height: AppSpacing.sm),
        DayBars(days: r.dias),
      ],
    );
  }
}
