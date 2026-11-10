import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:tiquetera/app.dart';
import 'package:tiquetera/core/config/env.dart';
import 'package:tiquetera/core/providers.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initializeDateFormatting('es_CO');
  final env = Env.fromDartDefines();
  runApp(
    ProviderScope(
      overrides: [envProvider.overrideWithValue(env)],
      child: const TiqueteraApp(),
    ),
  );
}
