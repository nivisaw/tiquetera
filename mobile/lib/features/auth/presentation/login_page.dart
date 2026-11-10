import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tiquetera/core/network/app_failure.dart';
import 'package:tiquetera/core/theme/app_theme.dart';
import 'package:tiquetera/features/auth/presentation/auth_controller.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class LoginPage extends ConsumerStatefulWidget {
  const new({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _password = TextEditingController();
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _username.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final l10n = AppLocalizations.of(context);
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await ref
          .read(authControllerProvider.notifier)
          .login(_username.text.trim(), _password.text);
    } on AppFailure catch (f) {
      if (mounted) setState(() => _error = loginErrorMessage(l10n, f));
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      l10n.appTitle,
                      style: Theme.of(context).textTheme.headlineMedium,
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    TextFormField(
                      key: const Key('login_username'),
                      controller: _username,
                      enabled: !_submitting,
                      autocorrect: false,
                      textInputAction: TextInputAction.next,
                      autofillHints: const [AutofillHints.username],
                      decoration: InputDecoration(labelText: l10n.loginUser),
                      validator: (v) =>
                          (v ?? '').trim().isEmpty ? l10n.fieldRequired : null,
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextFormField(
                      key: const Key('login_password'),
                      controller: _password,
                      enabled: !_submitting,
                      obscureText: true,
                      textInputAction: TextInputAction.done,
                      autofillHints: const [AutofillHints.password],
                      decoration: InputDecoration(
                        labelText: l10n.loginPassword,
                      ),
                      validator: (v) =>
                          (v ?? '').isEmpty ? l10n.fieldRequired : null,
                      onFieldSubmitted: (_) => _submitting ? null : _submit(),
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: AppSpacing.md),
                      _ErrorBanner(message: _error!),
                    ],
                    const SizedBox(height: AppSpacing.lg),
                    FilledButton(
                      key: const Key('login_submit'),
                      onPressed: _submitting ? null : _submit,
                      child: _submitting
                          ? SizedBox.square(
                              dimension: 24,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                semanticsLabel: l10n.loading,
                              ),
                            )
                          : Text(l10n.loginSubmit),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

String loginErrorMessage(AppLocalizations l10n, AppFailure f) => switch (f) {
  UnauthorizedFailure() => l10n.loginInvalidCredentials,
  ApiFailure(code: 'RATE_LIMITED') => l10n.loginRateLimited,
  NetworkFailure() => l10n.errorNetwork,
  TimeoutFailure() => l10n.errorTimeout,
  ApiFailure() || UnknownFailure() => l10n.errorUnknown,
};

class _ErrorBanner extends StatelessWidget {
  const new({required this.message});

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
