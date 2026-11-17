import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/features/escaner/domain/consumo.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

/// Pantalla completa de éxito o rechazo; vuelve al escáner en
/// [autoReturn] o al tocarla.
class ResultadoView extends StatefulWidget {
  const new({
    required this.onDone,
    this.consumo,
    this.motivo,
    this.autoReturn = const Duration(seconds: 3),
    super.key,
  }) : assert((consumo == null) != (motivo == null), 'consumo xor motivo');

  final ConsumoOk? consumo;
  final RechazoMotivo? motivo;
  final VoidCallback onDone;
  final Duration autoReturn;

  @override
  State<ResultadoView> createState() => _ResultadoViewState();
}

class _ResultadoViewState extends State<ResultadoView> {
  Timer? _timer;

  bool get _ok => widget.consumo != null;

  @override
  void initState() {
    super.initState();
    _timer = Timer(widget.autoReturn, _done);
    unawaited(_feedback());
  }

  Future<void> _feedback() async {
    if (_ok) {
      await HapticFeedback.mediumImpact();
      await SystemSound.play(SystemSoundType.click);
      return;
    }
    await HapticFeedback.heavyImpact();
    await Future<void>.delayed(const Duration(milliseconds: 180));
    await HapticFeedback.heavyImpact();
  }

  void _done() {
    if (_timer == null) return;
    _timer!.cancel();
    _timer = null;
    widget.onDone();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final consumo = widget.consumo;
    return Material(
      color: _ok ? AppColors.success : AppColors.rejection,
      child: InkWell(
        onTap: _done,
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: DefaultTextStyle.merge(
                style: const TextStyle(color: Colors.white),
                textAlign: TextAlign.center,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    _ResultIcon(ok: _ok),
                    const SizedBox(height: AppSpacing.md),
                    if (consumo != null)
                      ..._exito(context, l10n, consumo)
                    else
                      ..._rechazo(context, l10n, widget.motivo!),
                    const SizedBox(height: AppSpacing.lg),
                    Text(l10n.escanerTocaParaVolver),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _exito(
    BuildContext context,
    AppLocalizations l10n,
    ConsumoOk c,
  ) {
    final text = Theme.of(context).textTheme;
    const white = TextStyle(color: Colors.white);
    return [
      Text(l10n.escanerExito, style: text.headlineMedium?.merge(white)),
      const SizedBox(height: AppSpacing.sm),
      Text(c.nombre, style: text.titleLarge?.merge(white)),
      const SizedBox(height: AppSpacing.md),
      Text(
        '${c.saldoRestante}',
        style: const TextStyle(
          color: Colors.white,
          fontSize: 72,
          fontWeight: FontWeight.bold,
        ),
      ),
      Text(l10n.escanerSaldoLabel, style: text.titleMedium?.merge(white)),
      const SizedBox(height: AppSpacing.sm),
      Text(
        l10n.escanerVence(ClienteRules.displayDate(c.fechaVencimiento)),
        style: text.titleMedium?.merge(white),
      ),
    ];
  }

  List<Widget> _rechazo(
    BuildContext context,
    AppLocalizations l10n,
    RechazoMotivo m,
  ) {
    final text = Theme.of(context).textTheme;
    const white = TextStyle(color: Colors.white);
    return [
      Text(l10n.escanerRechazoTitle, style: text.headlineMedium?.merge(white)),
      const SizedBox(height: AppSpacing.md),
      Text(rechazoMessage(l10n, m), style: text.titleLarge?.merge(white)),
    ];
  }
}

String rechazoMessage(AppLocalizations l10n, RechazoMotivo m) => switch (m) {
  RechazoMotivo.qrInvalido => l10n.rechazoQrInvalido,
  RechazoMotivo.clienteNoExiste => l10n.rechazoClienteNoExiste,
  RechazoMotivo.clienteInactivo => l10n.rechazoClienteInactivo,
  RechazoMotivo.sinTiquetera => l10n.rechazoSinTiquetera,
  RechazoMotivo.tiqueteraVencida => l10n.rechazoTiqueteraVencida,
  RechazoMotivo.sinSaldo => l10n.rechazoSinSaldo,
  RechazoMotivo.sinRed => l10n.rechazoSinRed,
  RechazoMotivo.demasiadosIntentos => l10n.rechazoDemasiados,
  RechazoMotivo.error => l10n.rechazoError,
};

class _ResultIcon extends StatelessWidget {
  const new({required this.ok});

  final bool ok;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0.3, end: 1),
      duration: const Duration(milliseconds: 350),
      curve: Curves.easeOutBack,
      builder: (_, scale, child) => Transform.scale(scale: scale, child: child),
      child: Icon(
        ok ? Icons.check_circle : Icons.cancel,
        size: 120,
        color: Colors.white,
        semanticLabel: ok ? l10n.escanerExito : l10n.escanerRechazoTitle,
      ),
    );
  }
}
