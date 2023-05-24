import 'package:flutter/material.dart';
import 'package:ui/service/api.dart';

import 'models/constants.dart';

class Landing extends StatefulWidget {
  const Landing({super.key});

  @override
  State<Landing> createState() => _LandingState();
}

class _LandingState extends State<Landing> {
  String _token = "";

  @override
  void initState() {
    super.initState();
    _loadUserInfo();
  }

  _loadUserInfo() async {
    _token = (await ClientAuthentication.loadToken() ?? "");
    _redirectLoginOrHome();
  }

  _redirectLoginOrHome() {
    if (_token == "") {
      Navigator.pushNamedAndRemoveUntil(
          context, Routes.login.route, (Route<dynamic> route) => false);
    } else {
      Navigator.pushNamedAndRemoveUntil(
          context, Routes.config.route, (Route<dynamic> route) => false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
