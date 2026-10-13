// This is a generated file - do not edit.
//
// Generated from checkout/v1/checkout.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class AmountOption extends $pb.ProtobufEnum {
  static const AmountOption AMOUNT_OPTION_FIXED_UNSPECIFIED = AmountOption._(
      0, _omitEnumNames ? '' : 'AMOUNT_OPTION_FIXED_UNSPECIFIED');
  static const AmountOption AMOUNT_OPTION_VARIABLE =
      AmountOption._(1, _omitEnumNames ? '' : 'AMOUNT_OPTION_VARIABLE');

  static const $core.List<AmountOption> values = <AmountOption>[
    AMOUNT_OPTION_FIXED_UNSPECIFIED,
    AMOUNT_OPTION_VARIABLE,
  ];

  static final $core.List<AmountOption?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static AmountOption? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AmountOption._(super.value, super.name);
}

class SessionStatus extends $pb.ProtobufEnum {
  static const SessionStatus SESSION_STATUS_PENDING_UNSPECIFIED =
      SessionStatus._(
          0, _omitEnumNames ? '' : 'SESSION_STATUS_PENDING_UNSPECIFIED');
  static const SessionStatus SESSION_STATUS_PROCESSING =
      SessionStatus._(1, _omitEnumNames ? '' : 'SESSION_STATUS_PROCESSING');
  static const SessionStatus SESSION_STATUS_COMPLETED =
      SessionStatus._(2, _omitEnumNames ? '' : 'SESSION_STATUS_COMPLETED');
  static const SessionStatus SESSION_STATUS_FAILED =
      SessionStatus._(3, _omitEnumNames ? '' : 'SESSION_STATUS_FAILED');
  static const SessionStatus SESSION_STATUS_EXPIRED =
      SessionStatus._(4, _omitEnumNames ? '' : 'SESSION_STATUS_EXPIRED');

  static const $core.List<SessionStatus> values = <SessionStatus>[
    SESSION_STATUS_PENDING_UNSPECIFIED,
    SESSION_STATUS_PROCESSING,
    SESSION_STATUS_COMPLETED,
    SESSION_STATUS_FAILED,
    SESSION_STATUS_EXPIRED,
  ];

  static final $core.List<SessionStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static SessionStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const SessionStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
