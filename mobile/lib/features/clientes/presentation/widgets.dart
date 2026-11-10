import 'package:flutter/material.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

String clienteErrorMessage(AppLocalizations l10n, Object f) => switch (f) {
  ApiFailure(code: 'TELEFONO_DUPLICADO') => l10n.errTelefonoDuplicado,
  ApiFailure(code: 'VALIDATION_ERROR') => l10n.errValidacion,
  ApiFailure(code: 'CLIENTE_INACTIVO') => l10n.errClienteInactivo,
  ApiFailure(code: 'NOT_FOUND') => l10n.errNotFound,
  UnauthorizedFailure() => l10n.errorUnauthorized,
  NetworkFailure() => l10n.errorNetwork,
  TimeoutFailure() => l10n.errorTimeout,
  _ => l10n.errorUnknown,
};

class ErrorView extends StatelessWidget {
  const new({required this.error, required this.onRetry, super.key});

  final Object error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.error_outline,
              size: 48,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(height: AppSpacing.md),
            Text(clienteErrorMessage(l10n, error), textAlign: TextAlign.center),
            const SizedBox(height: AppSpacing.md),
            FilledButton.tonal(onPressed: onRetry, child: Text(l10n.retry)),
          ],
        ),
      ),
    );
  }
}

class ErrorBanner extends StatelessWidget {
  const new({required this.message, super.key});

  final String message;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Semantics(
      liveRegion: true,
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.md),
        decoration: BoxDecoration(
          color: scheme.errorContainer,
          borderRadius: BorderRadius.circular(AppRadius.sm),
        ),
        child: Row(
          children: [
            Icon(Icons.error_outline, color: scheme.onErrorContainer),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Text(
                message,
                style: TextStyle(color: scheme.onErrorContainer),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Selector de almuerzos con botones −/+ y campo numérico (1..500).
class AlmuerzosField extends StatefulWidget {
  const new({
    required this.value,
    required this.onChanged,
    this.enabled = true,
    super.key,
  });

  final int value;
  final ValueChanged<int> onChanged;
  final bool enabled;

  @override
  State<AlmuerzosField> createState() => _AlmuerzosFieldState();
}

class _AlmuerzosFieldState extends State<AlmuerzosField> {
  late final _controller = TextEditingController(text: '${widget.value}');

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _step(int delta) {
    final next = ClienteRules.clampAlmuerzos(widget.value + delta);
    _controller.text = '$next';
    widget.onChanged(next);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final v = widget.value;
    return Row(
      children: [
        IconButton.outlined(
          key: const Key('almuerzos_minus'),
          tooltip: l10n.clienteDecrease,
          onPressed: widget.enabled && v > ClienteRules.minAlmuerzos
              ? () => _step(-1)
              : null,
          icon: const Icon(Icons.remove),
        ),
        const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: TextFormField(
            key: const Key('cliente_almuerzos'),
            controller: _controller,
            enabled: widget.enabled,
            keyboardType: TextInputType.number,
            textAlign: TextAlign.center,
            decoration: InputDecoration(labelText: l10n.clienteAlmuerzos),
            validator: (s) {
              final n = int.tryParse(s ?? '');
              return n == null ||
                      n < ClienteRules.minAlmuerzos ||
                      n > ClienteRules.maxAlmuerzos
                  ? l10n.errAlmuerzos
                  : null;
            },
            onChanged: (s) {
              final n = int.tryParse(s);
              if (n != null) widget.onChanged(n);
            },
          ),
        ),
        const SizedBox(width: AppSpacing.sm),
        IconButton.outlined(
          key: const Key('almuerzos_plus'),
          tooltip: l10n.clienteIncrease,
          onPressed: widget.enabled && v < ClienteRules.maxAlmuerzos
              ? () => _step(1)
              : null,
          icon: const Icon(Icons.add),
        ),
      ],
    );
  }
}

/// Campo de fecha de vencimiento (no permite fechas anteriores a `today`).
class VencimientoField extends StatelessWidget {
  const new({
    required this.value,
    required this.today,
    required this.onChanged,
    this.enabled = true,
    super.key,
  });

  final DateTime value;
  final DateTime today;
  final ValueChanged<DateTime> onChanged;
  final bool enabled;

  Future<void> _pick(BuildContext context) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: value.isBefore(today) ? today : value,
      firstDate: today,
      lastDate: DateTime(today.year + 2, today.month, today.day),
    );
    if (picked != null) onChanged(picked);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Semantics(
      button: true,
      label: l10n.clienteSelectDate,
      child: InkWell(
        key: const Key('cliente_vencimiento'),
        onTap: enabled ? () => _pick(context) : null,
        borderRadius: BorderRadius.circular(AppRadius.sm),
        child: InputDecorator(
          decoration: InputDecoration(
            labelText: l10n.clienteVencimiento,
            suffixIcon: const Icon(Icons.calendar_today),
            enabled: enabled,
          ),
          child: Text(ClienteRules.displayDate(value)),
        ),
      ),
    );
  }
}
