import 'package:flutter/material.dart';
import 'package:tiquetera/features/auth/presentation/logout_button.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class EscanerPage extends StatelessWidget {
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.escanerTitle),
        actions: const [LogoutButton()],
      ),
    );
  }
}
