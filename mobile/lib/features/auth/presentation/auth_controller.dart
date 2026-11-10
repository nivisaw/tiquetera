import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:tiquetera/core/providers.dart';
import 'package:tiquetera/features/auth/data/auth_repository.dart';
import 'package:tiquetera/features/auth/domain/user.dart';

part 'auth_controller.g.dart';

@Riverpod(keepAlive: true)
AuthRepository authRepository(Ref ref) => AuthRepository(
  ref.watch(apiClientProvider),
  ref.watch(tokenStorageProvider),
);

/// Sesión actual (`null` = sin sesión), restaurada del almacenamiento seguro.
@Riverpod(keepAlive: true)
class AuthController extends _$AuthController {
  @override
  Future<User?> build() {
    ref.listen(sessionExpiredProvider, (_, _) => state = const AsyncData(null));
    return ref.read(authRepositoryProvider).restore();
  }

  Future<User> login(String username, String password) async {
    final user = await ref
        .read(authRepositoryProvider)
        .login(username: username, password: password);
    state = AsyncData(user);
    return user;
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    state = const AsyncData(null);
  }
}
