import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/clientes/presentation/clientes_providers.dart';
import 'package:tiquetera/features/clientes/presentation/qr_page.dart';
import 'package:tiquetera/features/clientes/presentation/widgets.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class ClienteDetallePage extends ConsumerWidget {
  const new({required this.id, super.key});

  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final detalle = ref.watch(clienteDetalleProvider(id));
    return Scaffold(
      appBar: AppBar(title: Text(l10n.detalleTitle)),
      body: switch (detalle) {
        AsyncData(:final value) => _Body(cliente: value),
        AsyncError(:final error) => ErrorView(
          error: error,
          onRetry: () => ref.invalidate(clienteDetalleProvider(id)),
        ),
        _ => Center(
          child: CircularProgressIndicator(semanticsLabel: l10n.loading),
        ),
      },
    );
  }
}

class _Body extends ConsumerStatefulWidget {
  const new({required this.cliente});

  final Cliente cliente;

  @override
  ConsumerState<_Body> createState() => _BodyState();
}

class _BodyState extends ConsumerState<_Body> {
  bool _busy = false;

  Future<void> _regenerar() async {
    final l10n = AppLocalizations.of(context);
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        icon: const Icon(Icons.warning_amber),
        title: Text(l10n.regenerarTitle),
        content: Text(l10n.regenerarBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            key: const Key('regenerar_confirm'),
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(l10n.regenerarConfirm),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    final c = widget.cliente;
    await _run(() async {
      final payload = await ref
          .read(clientesRepositoryProvider)
          .regenerarQr(c.id);
      if (!mounted) return;
      final t = c.tiquetera;
      await context.push(
        '/clientes/qr',
        extra: QrArgs(
          clienteId: c.id,
          nombre: c.nombre,
          saldo: t?.saldo ?? 0,
          fechaVencimiento: t?.fechaVencimiento ?? _today(),
          payload: payload,
        ),
      );
    });
  }

  Future<void> _recargar() async {
    final c = widget.cliente;
    final done = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (_) => RecargaSheet(clienteId: c.id, today: _today()),
    );
    if (done == true && mounted) {
      ref
        ..invalidate(clienteDetalleProvider(c.id))
        ..invalidate(clientesListProvider);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).recargaOk)),
      );
    }
  }

  DateTime _today() => ClienteRules.todayBogota(ref.read(clockProvider)());

  Future<void> _run(Future<void> Function() fn) async {
    final l10n = AppLocalizations.of(context);
    setState(() => _busy = true);
    try {
      await fn();
    } on AppFailure catch (f) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(clienteErrorMessage(l10n, f))));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final c = widget.cliente;
    final t = c.tiquetera;
    final text = Theme.of(context).textTheme;
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(c.nombre, style: text.headlineSmall),
        const SizedBox(height: AppSpacing.xs),
        Text(ClienteRules.displayTelefono(c.telefono), style: text.bodyLarge),
        if (c.documento != null) Text(c.documento!),
        if (c.email != null) Text(c.email!),
        if (!c.activo) ...[
          const SizedBox(height: AppSpacing.sm),
          Align(
            alignment: Alignment.centerLeft,
            child: Chip(label: Text(l10n.clienteInactivo)),
          ),
        ],
        const SizedBox(height: AppSpacing.lg),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: t == null
                ? Text(l10n.detalleSinTiquetera)
                : Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(l10n.saldoLabel(t.saldo), style: text.titleLarge),
                      Text(
                        l10n.venceLabel(
                          ClienteRules.displayDate(t.fechaVencimiento),
                        ),
                      ),
                    ],
                  ),
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        FilledButton.icon(
          key: const Key('detalle_recargar'),
          onPressed: _busy || !c.activo ? null : _recargar,
          icon: const Icon(Icons.add_card),
          label: Text(l10n.detalleRecargar),
        ),
        const SizedBox(height: AppSpacing.sm),
        OutlinedButton.icon(
          key: const Key('detalle_regenerar'),
          onPressed: _busy ? null : _regenerar,
          icon: const Icon(Icons.qr_code),
          label: Text(l10n.detalleRegenerar),
        ),
      ],
    );
  }
}

class RecargaSheet extends ConsumerStatefulWidget {
  const new({required this.clienteId, required this.today, super.key});

  final String clienteId;
  final DateTime today;

  @override
  ConsumerState<RecargaSheet> createState() => _RecargaSheetState();
}

class _RecargaSheetState extends ConsumerState<RecargaSheet> {
  final _formKey = GlobalKey<FormState>();
  int _almuerzos = 20;
  late DateTime _vence = widget.today.add(const Duration(days: 30));
  bool _submitting = false;
  String? _error;

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final l10n = AppLocalizations.of(context);
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await ref
          .read(clientesRepositoryProvider)
          .recargar(
            widget.clienteId,
            almuerzos: _almuerzos,
            fechaVencimiento: _vence,
          );
      if (mounted) Navigator.pop(context, true);
    } on AppFailure catch (f) {
      if (mounted) setState(() => _error = clienteErrorMessage(l10n, f));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: EdgeInsets.fromLTRB(
        AppSpacing.md,
        0,
        AppSpacing.md,
        AppSpacing.md + MediaQuery.viewInsetsOf(context).bottom,
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.recargaTitle,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            Text(l10n.recargaHint),
            const SizedBox(height: AppSpacing.md),
            AlmuerzosField(
              value: _almuerzos,
              enabled: !_submitting,
              onChanged: (v) => setState(() => _almuerzos = v),
            ),
            const SizedBox(height: AppSpacing.md),
            VencimientoField(
              value: _vence,
              today: widget.today,
              enabled: !_submitting,
              onChanged: (d) => setState(() => _vence = d),
            ),
            if (_error != null) ...[
              const SizedBox(height: AppSpacing.md),
              ErrorBanner(message: _error!),
            ],
            const SizedBox(height: AppSpacing.lg),
            FilledButton(
              key: const Key('recarga_save'),
              onPressed: _submitting ? null : _submit,
              child: Text(l10n.clienteSave),
            ),
          ],
        ),
      ),
    );
  }
}
