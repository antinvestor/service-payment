import { Namespace, Context } from "@ory/keto-namespace-types"

class profile_user implements Namespace {}

class tenancy_access implements Namespace {
  related: {
    member: (profile_user | tenancy_access)[]
    service: profile_user[]
  }
}

class service_ledger implements Namespace {
  related: {
    owner: profile_user[]
    admin: profile_user[]
    operator: profile_user[]
    viewer: profile_user[]
    member: profile_user[]
    service: (profile_user | tenancy_access)[]

    // Direct permission grants (accept service_ledger subject sets for service role bridging)
    manage_ledger: (profile_user | service_ledger)[]
    view_ledger: (profile_user | service_ledger)[]
    manage_account: (profile_user | service_ledger)[]
    view_account: (profile_user | service_ledger)[]
    create_transaction: (profile_user | service_ledger)[]
    reverse_transaction: (profile_user | service_ledger)[]
    update_transaction: (profile_user | service_ledger)[]
    view_transaction: (profile_user | service_ledger)[]
  }

  permits = {
    manage_ledger: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.manage_ledger.includes(ctx.subject),

    view_ledger: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.permits.manage_ledger(ctx) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.view_ledger.includes(ctx.subject),

    manage_account: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.manage_account.includes(ctx.subject),

    view_account: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.permits.manage_account(ctx) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.view_account.includes(ctx.subject),

    create_transaction: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.create_transaction.includes(ctx.subject),

    reverse_transaction: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.reverse_transaction.includes(ctx.subject),

    update_transaction: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.update_transaction.includes(ctx.subject),

    view_transaction: (ctx: Context): boolean =>
      this.related.service.includes(ctx.subject) ||
      this.related.owner.includes(ctx.subject) ||
      this.related.admin.includes(ctx.subject) ||
      this.related.operator.includes(ctx.subject) ||
      this.related.viewer.includes(ctx.subject) ||
      this.related.member.includes(ctx.subject) ||
      this.related.view_transaction.includes(ctx.subject),
  }
}
