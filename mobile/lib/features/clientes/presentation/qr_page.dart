import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/clientes/domain/cliente_rules.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

/// Datos que muestra la pantalla QR. El payload no se persiste: solo llega
/// al crear el cliente o regenerar el QR.
class QrArgs {
  const new({
    required this.clienteId,
    required this.nombre,
    required this.saldo,
    required this.fechaVencimiento,
    required this.payload,
  });

  final String clienteId;
  final String nombre;
  final int saldo;
  final DateTime fechaVencimiento;
  final String payload;
}

class QrPage extends StatelessWidget {
  const new({required this.args, super.key});

  final QrArgs args;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.qrTitle)),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Center(
                child: QrCard(payload: args.payload, nombre: args.nombre),
              ),
              const SizedBox(height: AppSpacing.lg),
              Text(
                args.nombre,
                style: text.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                l10n.saldoLabel(args.saldo),
                style: text.titleLarge,
                textAlign: TextAlign.center,
              ),
              Text(
                l10n.venceLabel(
                  ClienteRules.displayDate(args.fechaVencimiento),
                ),
                style: text.bodyLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                key: const Key('qr_done'),
                onPressed: () => context.go('/clientes/${args.clienteId}'),
                child: Text(l10n.qrDone),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// QR negro sobre blanco con margen (zona de silencio), siempre en claro
/// para que se lea también en modo oscuro.
class QrCard extends StatelessWidget {
  const new({required this.payload, required this.nombre, super.key});

  final String payload;
  final String nombre;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    // Resta el padding de la página y el del recuadro blanco.
    final side = (MediaQuery.sizeOf(context).shortestSide - 4 * AppSpacing.lg)
        .clamp(160.0, 360.0);
    return Semantics(
      image: true,
      label: l10n.qrSemantics(nombre),
      child: ExcludeSemantics(
        child: Container(
          key: const Key('qr_card'),
          color: Colors.white,
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: QrImageView(
            data: payload,
            size: side,
            padding: EdgeInsets.zero,
            backgroundColor: Colors.white,
          ),
        ),
      ),
    );
  }
}
