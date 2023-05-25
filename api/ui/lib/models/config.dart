enum ConfigProperty {
  enableApi(name: 'user.enableApi'),
  isAdmin(name: 'user.isAdmin'),
  currency(name: 'user.currency'),
  vacationTag(name: 'user.vacationTag'),
  timezoneOffset(name: 'user.tzOffset'),
  omitLeadingSlash(name: 'user.omitCommandSlash');

  const ConfigProperty({required this.name});

  final String name;
}

class Config {
  final bool enableApi;
  final bool isAdmin;
  String? currency;
  String? vacationTag;
  int timezoneOffset;
  bool omitLeadingCommandSlash;

  Config(this.enableApi, this.isAdmin, this.currency, this.vacationTag,
      this.timezoneOffset, this.omitLeadingCommandSlash);

  Map<String, dynamic> diffChanged(Config cnf) {
    Map<String, dynamic> diff = {};
    if (currency != cnf.currency) {
      diff[ConfigProperty.currency.name] = cnf.currency;
    }
    if (vacationTag != cnf.vacationTag) {
      diff[ConfigProperty.vacationTag.name] = cnf.vacationTag;
    }
    if (timezoneOffset != cnf.timezoneOffset) {
      diff[ConfigProperty.timezoneOffset.name] = cnf.timezoneOffset;
    }
    if (omitLeadingCommandSlash != cnf.omitLeadingCommandSlash) {
      diff[ConfigProperty.omitLeadingSlash.name] = cnf.omitLeadingCommandSlash;
    }
    return diff;
  }
}
