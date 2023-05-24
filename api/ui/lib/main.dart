import 'package:flutter/material.dart';
import 'package:ui/screens/home.dart';
import 'package:ui/screens/login.dart';
import 'package:go_router/go_router.dart';

import 'models/constants.dart';

final _router = GoRouter(
  routes: [
    GoRoute(
      path: Routes.root.route,
      builder: (context, state) => const HomePage(),
    ),
    GoRoute(
      path: Routes.login.route,
      builder: (BuildContext context, GoRouterState state) {
        return const LoginPage();
      },
    ),
  ],
);

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Beancount-Bot-Tg',
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
      ),
      routerConfig: _router,
    );
  }
}
