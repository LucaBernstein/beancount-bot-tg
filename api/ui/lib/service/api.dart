import 'dart:io';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import '../models/config.dart';
import '../models/constants.dart';

class BaseCrud {
  static String baseUrl() {
    if (kDebugMode) {
      return 'http://localhost:8080';
    }
    return '';
  }
}

class ClientAuthentication extends BaseCrud {
  String? userId;
  String? token;

  Future<bool> loadExistingToken() async {
    if (token != null && token!.isNotEmpty) {
      return true;
    }
    token = await loadToken();
    return token != null && token!.isNotEmpty;
  }

  Future<(String? error,)> generateVerificationCode(String userId) async {
    this.userId = userId;
    String url = '${BaseCrud.baseUrl()}/api/token/verification/$userId';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    if (response.statusCode == 204) { return (null,);} // OPTIONS
    if (response.statusCode == 201) {
      return (null,);
    }
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    return ('${responseMap['error']} (${response.statusCode})',);
  }

  Future<(String? token, String? errorMsg)> validateVerificationCode(
      String verification) async {
    String url = '${BaseCrud.baseUrl()}/api/token/grant/$userId/$verification';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    if (response.statusCode == 204) { return (null, null);} // OPTIONS
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      token = responseMap['token'];
      await storeToken(token!);
      return (token, null);
    }
    return (null, '${responseMap['error']} (${response.statusCode})');
  }

  Future<int> revokeAccess() async {
    String url = '${BaseCrud.baseUrl()}/api/token/revoke/$token';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    int affected = responseMap['affected'];

    // Cleanup local stores
    await storeToken("");
    token = null;
    userId = null;

    return affected;
  }

  static storeToken(String token) async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    prefs.setString(KeyToken, token);
  }
  static Future<String?> loadToken() async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(KeyToken);
  }

  Future<(Config? config, String? errorMsg)> getConfig() async {
    String url = '${BaseCrud.baseUrl()}/api/config/';
    final response = await http.get(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return (Config(
          responseMap['user.enableApi'],
          responseMap['user.isAdmin'] ?? false,
          responseMap['user.currency'],
          responseMap['user.vacationTag'],
          responseMap['user.tzOffset'],
          responseMap['user.omitCommandSlash'],
          ), null);
    }
    return (null, '${responseMap['error']} (${response.statusCode})');
  }
}
