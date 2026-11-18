import 'package:flutter/material.dart';

abstract final class AppSpacing {
  static const double xs = 4;
  static const double sm = 8;
  static const double md = 16;
  static const double lg = 24;
}

abstract final class AppRadius {
  static const double sm = 8;
  static const double md = 12;
  static const double lg = 16;
}

/// Paleta: azul marino (confianza, marca) + ámbar (acento cálido de
/// restaurante) sobre neutros pizarra. Éxito/rechazo del escáner fijos
/// (rules/ux-ui.md). Todos los pares texto/fondo cumplen WCAG AA.
abstract final class AppColors {
  static const Color navy = Color(0xFF1E3A8A);
  static const Color amber = Color(0xFFB45309);
  static const Color success = Color(0xFF1B873F);
  static const Color rejection = Color(0xFFC62828);

  static final ColorScheme light = ColorScheme.fromSeed(seedColor: navy)
      .copyWith(
        primary: navy,
        onPrimary: Colors.white,
        primaryContainer: const Color(0xFFDBE4FF),
        onPrimaryContainer: const Color(0xFF0B1F5C),
        secondary: amber,
        onSecondary: Colors.white,
        secondaryContainer: const Color(0xFFFFEDD5),
        onSecondaryContainer: const Color(0xFF5A2A00),
        tertiary: const Color(0xFF0F766E),
        onTertiary: Colors.white,
        error: const Color(0xFFB91C1C),
        onError: Colors.white,
        surface: const Color(0xFFF8FAFC),
        onSurface: const Color(0xFF0F172A),
        onSurfaceVariant: const Color(0xFF475569),
        outline: const Color(0xFF94A3B8),
        outlineVariant: const Color(0xFFE2E8F0),
        surfaceContainerLowest: Colors.white,
        surfaceContainerLow: const Color(0xFFF1F5F9),
        surfaceContainer: const Color(0xFFEEF2F7),
        surfaceContainerHigh: const Color(0xFFE2E8F0),
        surfaceContainerHighest: const Color(0xFFDCE3EC),
        surfaceTint: Colors.transparent,
      );

  static final ColorScheme dark =
      ColorScheme.fromSeed(
        seedColor: navy,
        brightness: Brightness.dark,
      ).copyWith(
        primary: const Color(0xFF9DB4FF),
        onPrimary: const Color(0xFF0B1F5C),
        primaryContainer: const Color(0xFF1E3A8A),
        onPrimaryContainer: const Color(0xFFDBE4FF),
        secondary: const Color(0xFFFBBF24),
        onSecondary: const Color(0xFF3B2300),
        secondaryContainer: const Color(0xFF7C3A06),
        onSecondaryContainer: const Color(0xFFFFEDD5),
        tertiary: const Color(0xFF5EEAD4),
        onTertiary: const Color(0xFF042F2E),
        error: const Color(0xFFF87171),
        onError: const Color(0xFF450A0A),
        surface: const Color(0xFF0B1220),
        onSurface: const Color(0xFFE2E8F0),
        onSurfaceVariant: const Color(0xFF94A3B8),
        outline: const Color(0xFF475569),
        outlineVariant: const Color(0xFF1E293B),
        surfaceContainerLowest: const Color(0xFF070D18),
        surfaceContainerLow: const Color(0xFF111A2B),
        surfaceContainer: const Color(0xFF151F32),
        surfaceContainerHigh: const Color(0xFF1B263B),
        surfaceContainerHighest: const Color(0xFF223047),
        surfaceTint: Colors.transparent,
      );
}

abstract final class AppTheme {
  static const double minTapTarget = 48;

  static ThemeData light() => _build(AppColors.light);
  static ThemeData dark() => _build(AppColors.dark);

  static ThemeData _build(ColorScheme scheme) {
    const shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.all(Radius.circular(AppRadius.md)),
    );
    const buttonSize = Size(minTapTarget, minTapTarget);
    const buttonText = TextStyle(fontWeight: FontWeight.w600, fontSize: 15);
    final base = ThemeData(colorScheme: scheme, useMaterial3: true);
    final text = base.textTheme.copyWith(
      headlineMedium: base.textTheme.headlineMedium?.copyWith(
        fontWeight: FontWeight.w700,
        letterSpacing: -0.5,
      ),
      headlineSmall: base.textTheme.headlineSmall?.copyWith(
        fontWeight: FontWeight.w700,
        letterSpacing: -0.25,
      ),
      titleLarge: base.textTheme.titleLarge?.copyWith(
        fontWeight: FontWeight.w600,
      ),
      titleMedium: base.textTheme.titleMedium?.copyWith(
        fontWeight: FontWeight.w600,
      ),
    );
    return base.copyWith(
      textTheme: text,
      scaffoldBackgroundColor: scheme.surface,
      materialTapTargetSize: MaterialTapTargetSize.padded,
      appBarTheme: AppBarTheme(
        backgroundColor: scheme.surface,
        foregroundColor: scheme.onSurface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 1,
        shadowColor: scheme.outlineVariant,
        titleTextStyle: text.titleLarge?.copyWith(color: scheme.onSurface),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          minimumSize: buttonSize,
          shape: shape,
          textStyle: buttonText,
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          minimumSize: buttonSize,
          shape: shape,
          textStyle: buttonText,
          side: BorderSide(color: scheme.outline),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          minimumSize: buttonSize,
          textStyle: buttonText,
        ),
      ),
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: scheme.secondary,
        foregroundColor: scheme.onSecondary,
        elevation: 2,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(AppRadius.lg)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: scheme.surfaceContainerLowest,
        border: _inputBorder(scheme.outline),
        enabledBorder: _inputBorder(scheme.outline),
        focusedBorder: _inputBorder(scheme.primary, width: 2),
        errorBorder: _inputBorder(scheme.error),
        focusedErrorBorder: _inputBorder(scheme.error, width: 2),
      ),
      cardTheme: CardThemeData(
        elevation: 0,
        color: scheme.surfaceContainerLowest,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: const BorderRadius.all(Radius.circular(AppRadius.lg)),
          side: BorderSide(color: scheme.outlineVariant),
        ),
      ),
      chipTheme: ChipThemeData(
        shape: const StadiumBorder(),
        side: BorderSide(color: scheme.outlineVariant),
        selectedColor: scheme.primaryContainer,
        labelStyle: text.labelLarge,
      ),
      listTileTheme: ListTileThemeData(
        iconColor: scheme.onSurfaceVariant,
        shape: shape,
      ),
      dividerTheme: DividerThemeData(color: scheme.outlineVariant, space: 1),
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape: shape,
        backgroundColor: scheme.onSurface,
        contentTextStyle: TextStyle(color: scheme.surface),
      ),
      progressIndicatorTheme: ProgressIndicatorThemeData(
        color: scheme.primary,
        linearTrackColor: scheme.surfaceContainerHigh,
      ),
      dialogTheme: const DialogThemeData(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(AppRadius.lg)),
        ),
      ),
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: scheme.surfaceContainerLowest,
        surfaceTintColor: Colors.transparent,
      ),
    );
  }

  static OutlineInputBorder _inputBorder(Color color, {double width = 1}) =>
      OutlineInputBorder(
        borderRadius: const BorderRadius.all(Radius.circular(AppRadius.md)),
        borderSide: BorderSide(color: color, width: width),
      );
}
