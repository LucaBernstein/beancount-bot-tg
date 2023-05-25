import 'dart:io';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:ui/models/transaction.dart';

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

  Future<(bool, String?)> loadExistingToken() async {
    if (token != null && token!.isNotEmpty) {
      return (true, token);
    }
    token = await loadToken();
    return (token != null && token!.isNotEmpty, token);
  }

  Future<(String? error,)> generateVerificationCode(String userId) async {
    this.userId = userId;
    String url = '${BaseCrud.baseUrl()}/api/token/verification/$userId';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    if (response.statusCode == 204) {
      return (null,);
    } // OPTIONS
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
    if (response.statusCode == 204) {
      return (null, null);
    } // OPTIONS
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
    prefs.setString(keyToken, token);
  }

  static Future<String?> loadToken() async {
    SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(keyToken);
  }

  Future<(Config? config, String? errorMsg)> getConfig() async {
    String url = '${BaseCrud.baseUrl()}/api/config/';
    final response = await http.get(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return (
        Config(
          responseMap['user.enableApi'],
          responseMap['user.isAdmin'] ?? false,
          responseMap['user.currency'],
          responseMap['user.vacationTag'],
          responseMap['user.tzOffset'],
          responseMap['user.omitCommandSlash'],
        ),
        null
      );
    }
    return (null, '${responseMap['error']} (${response.statusCode})');
  }

  Future<(String? errorMsg,)> saveConfig(Config cnf) async {
    Config? priorConfig;
    String? retrievalError;
    (priorConfig, retrievalError) = await getConfig();
    if (retrievalError != null && retrievalError.isNotEmpty) {
      return (retrievalError,);
    }
    String url = '${BaseCrud.baseUrl()}/api/config/';
    List<String> successful = [];
    for (var e in priorConfig!.diffChanged(cnf).entries) {
      if (e.key == '') {
        continue;
      }
      final response = await http.post(Uri.parse(url),
          headers: <String, String>{
            HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
            HttpHeaders.authorizationHeader: 'Bearer $token',
          },
          body: jsonEncode({'setting': e.key, 'value': e.value}));
      if (response.statusCode == 200) {
        successful.add(e.key);
      } else {
        return (
          "Failed at saving setting ''. Updated successfully so far: [${successful.join(', ')}]",
        );
      }
    }
    return (null,);
  }

  Future<(Map<String, List<String>>? suggestions, String? errorMsg)>
      getSuggestions() async {
    String url = '${BaseCrud.baseUrl()}/api/suggestions/list';
    final response = await http.get(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    Map<String, List<String>> suggestions = {};
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      for (var e in responseMap.entries) {
        suggestions[e.key] = [];
        for (var s in e.value) {
          suggestions[e.key]!.add(s);
        }
      }
      return (suggestions, null);
    }
    return (null, '${responseMap['error']} (${response.statusCode})');
  }

  Future<(String? errorMsg,)> deleteSuggestion(
      String type, String? value) async {
    String url =
        '${BaseCrud.baseUrl()}/api/suggestions/list/$type/${value ?? ''}';
    final response =
        await http.delete(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return (null,);
    }
    return ('${responseMap['error']} (${response.statusCode})',);
  }

  Future<(List<Transaction> tx, String? errorMsg)> getTransactions() async {
    String url = '${BaseCrud.baseUrl()}/api/transactions/list';
    final response = await http.get(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    List<Transaction> transactions = [];
    List<dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      for (var e in responseMap) {
        transactions.add(Transaction(
            id: e['id'],
            createdAt: e['createdAt'],
            booking: e['booking'],
            isArchived: e['isArchived']));
      }
      return (transactions, null);
    }
    return (
      transactions,
      'Error retrieving transactions (${response.statusCode})'
    );
  }

  Future<(String? errorMsg,)> deleteTransaction(int id) async {
    String url = '${BaseCrud.baseUrl()}/api/transactions/list/$id';
    final response =
        await http.delete(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
      HttpHeaders.authorizationHeader: 'Bearer $token',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return (null,);
    }
    return ('${responseMap['error']} (${response.statusCode})',);
  }
}
