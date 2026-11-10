import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:tiquetera/features/auth/presentation/auth_controller.dart';
import 'package:tiquetera/l10n/gen/app_localizations.dart';

class LogoutButton extends ConsumerWidget {
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => IconButton(
    tooltip: AppLocalizations.of(context).logout,
    icon: const Icon(Icons.logout),
    onPressed: () => ref.read(authControllerProvider.notifier).logout(),
  );
}
