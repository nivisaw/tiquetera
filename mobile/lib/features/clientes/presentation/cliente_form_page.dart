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

class ClienteFormPage extends ConsumerStatefulWidget {
  const new({super.key});

  static const int defaultAlmuerzos = 20;
  static const int defaultDias = 30;

  @override
  ConsumerState<ClienteFormPage> createState() => _ClienteFormPageState();
}

class _ClienteFormPageState extends ConsumerState<ClienteFormPage> {
  final _formKey = GlobalKey<FormState>();
  final _nombre = TextEditingController();
  final _telefono = TextEditingController();
  final _documento = TextEditingController();
  final _email = TextEditingController();
  int _almuerzos = ClienteFormPage.defaultAlmuerzos;
  late final DateTime _today = ClienteRules.todayBogota(
    ref.read(clockProvider)(),
  );
  late DateTime _vence = _today.add(
    const Duration(days: ClienteFormPage.defaultDias),
  );
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _nombre.dispose();
    _telefono.dispose();
    _documento.dispose();
    _email.dispose();
    super.dispose();
  }

  String? _opt(TextEditingController c) =>
      c.text.trim().isEmpty ? null : c.text.trim();

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final l10n = AppLocalizations.of(context);
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      final r = await ref
          .read(clientesRepositoryProvider)
          .create(
            NuevoCliente(
              nombre: _nombre.text.trim(),
              telefono: ClienteRules.toE164(_telefono.text),
              documento: _opt(_documento),
              email: _opt(_email),
              almuerzos: _almuerzos,
              fechaVencimiento: _vence,
            ),
          );
      ref.invalidate(clientesListProvider);
      if (!mounted) return;
      context.pushReplacement(
        '/clientes/qr',
        extra: QrArgs(
          clienteId: r.cliente.id,
          nombre: r.cliente.nombre,
          saldo: r.tiquetera.saldo,
          fechaVencimiento: r.tiquetera.fechaVencimiento,
          payload: r.qrPayload,
        ),
      );
    } on AppFailure catch (f) {
      if (mounted) setState(() => _error = clienteErrorMessage(l10n, f));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.clientesNew)),
      body: SafeArea(
        child: Form(
          key: _formKey,
          autovalidateMode: AutovalidateMode.onUserInteraction,
          child: ListView(
            padding: const EdgeInsets.all(AppSpacing.md),
            children: [
              ..._fields(l10n),
              if (_error != null) ...[
                const SizedBox(height: AppSpacing.md),
                ErrorBanner(message: _error!),
              ],
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                key: const Key('cliente_save'),
                onPressed: _submitting ? null : _submit,
                child: _submitting
                    ? SizedBox.square(
                        dimension: 24,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          semanticsLabel: l10n.loading,
                        ),
                      )
                    : Text(l10n.clienteSave),
              ),
            ],
          ),
        ),
      ),
    );
  }

  List<Widget> _fields(AppLocalizations l10n) {
    const gap = SizedBox(height: AppSpacing.md);
    final on = !_submitting;
    return [
      TextFormField(
        key: const Key('cliente_nombre'),
        controller: _nombre,
        enabled: on,
        textCapitalization: TextCapitalization.words,
        textInputAction: TextInputAction.next,
        maxLength: ClienteRules.nombreMax,
        decoration: InputDecoration(labelText: l10n.clienteNombre),
        validator: (v) =>
            ClienteRules.validNombre(v ?? '') ? null : l10n.errNombre,
      ),
      gap,
      TextFormField(
        key: const Key('cliente_telefono'),
        controller: _telefono,
        enabled: on,
        keyboardType: TextInputType.phone,
        textInputAction: TextInputAction.next,
        autofillHints: const [AutofillHints.telephoneNumberNational],
        decoration: InputDecoration(
          labelText: l10n.clienteTelefono,
          prefixText: '${ClienteRules.countryPrefix} ',
        ),
        validator: (v) =>
            ClienteRules.validTelefono(v ?? '') ? null : l10n.errTelefono,
      ),
      gap,
      TextFormField(
        key: const Key('cliente_documento'),
        controller: _documento,
        enabled: on,
        textInputAction: TextInputAction.next,
        decoration: InputDecoration(labelText: l10n.clienteDocumento),
        validator: (v) =>
            ClienteRules.validDocumento(v ?? '') ? null : l10n.errDocumento,
      ),
      gap,
      TextFormField(
        key: const Key('cliente_email'),
        controller: _email,
        enabled: on,
        keyboardType: TextInputType.emailAddress,
        autocorrect: false,
        decoration: InputDecoration(labelText: l10n.clienteEmail),
        validator: (v) =>
            ClienteRules.validEmail(v ?? '') ? null : l10n.errEmail,
      ),
      gap,
      AlmuerzosField(
        value: _almuerzos,
        enabled: on,
        onChanged: (v) => setState(() => _almuerzos = v),
      ),
      gap,
      VencimientoField(
        value: _vence,
        today: _today,
        enabled: on,
        onChanged: (d) => setState(() => _vence = d),
      ),
    ];
  }
}
