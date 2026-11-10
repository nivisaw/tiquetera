import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/features/home/presentation/home_page.dart';

part 'app_router.g.dart';

abstract final class AppRoutes {
  static const home = '/';
}

@Riverpod(keepAlive: true)
GoRouter appRouter(Ref ref) => GoRouter(
  routes: [
    GoRoute(
      path: AppRoutes.home,
      builder: (context, state) => const HomePage(),
    ),
  ],
);
