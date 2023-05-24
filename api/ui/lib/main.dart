import 'package:flutter/material.dart';
import 'package:ui/screens/config.dart';
import 'package:ui/screens/login.dart';

import 'landing.dart';
import 'models/constants.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Beancount-Bot-Tg',
      routes: {
        Routes.root.route: (context) => const Landing(),
        Routes.login.route: (context) => const LoginPage(),
        Routes.config.route: (context) => const ConfigPage(),
      },
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
      ),
    );
  }
}
