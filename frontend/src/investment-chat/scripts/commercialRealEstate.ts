import type { ChatAnswers, ChatScript, ChatStep } from '../types';

function choiceLabel(step: ChatStep, value: unknown): string {
  return step.choices?.find((choice) => choice.value === value)?.label || String(value || '-');
}

function tenantCount(answers: ChatAnswers): number {
  const value = Number(answers.tenant_count || 0);
  return Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0;
}

function leaseCount(answers: ChatAnswers, tenantIndex: number): number {
  const value = Number(answers[`tenant_${tenantIndex}_lease_count`] || 0);
  return Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0;
}

function tenantNameStep(tenantIndex: number): ChatStep {
  return {
    id: `tenant_${tenantIndex}_name`,
    prompt: `What is the name or label for tenant ${tenantIndex}?`,
    input: 'text',
  };
}

function tenantLeaseCountStep(tenantIndex: number): ChatStep {
  return {
    id: `tenant_${tenantIndex}_lease_count`,
    prompt: `How many leases or lease phases should we capture for tenant ${tenantIndex}?`,
    input: 'number',
  };
}

function leaseFieldStep(tenantIndex: number, leaseIndex: number, field: string): ChatStep {
  const baseId = `tenant_${tenantIndex}_lease_${leaseIndex}_${field}`;
  switch (field) {
    case 'suite':
      return {
        id: baseId,
        prompt: `For tenant ${tenantIndex}, lease ${leaseIndex}, what suite or space does this lease cover?`,
        input: 'text',
      };
    case 'lease_type':
      return {
        id: baseId,
        prompt: `What lease type applies to tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'choice',
        choices: [
          { label: 'Gross', value: 'gross' },
          { label: 'Modified gross', value: 'modified_gross' },
          { label: 'Net', value: 'net' },
          { label: 'NNN', value: 'nnn' },
          { label: 'Percentage rent', value: 'percentage_rent' },
          { label: 'Other', value: 'other' },
        ],
      };
    case 'start':
      return {
        id: baseId,
        prompt: `What is the start date for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'date',
      };
    case 'end':
      return {
        id: baseId,
        prompt: `What is the end date for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'date',
      };
    case 'term_notes':
      return {
        id: baseId,
        prompt: `What term structure or lease timing details matter for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'textarea',
        helpText: 'Include free rent, phased occupancy, break options, or unusual term mechanics.',
      };
    case 'base_rent':
      return {
        id: baseId,
        prompt: `What is the base rent for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'number',
      };
    case 'rent_growth':
      return {
        id: baseId,
        prompt: `How does rent change over time for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'textarea',
        helpText: 'For example annual bumps, CPI linkage, step-ups, or fixed resets.',
      };
    case 'revenue_share':
      return {
        id: baseId,
        prompt: `Does tenant ${tenantIndex}, lease ${leaseIndex} include a revenue share agreement?`,
        input: 'yesno',
      };
    case 'revenue_share_terms':
      return {
        id: baseId,
        prompt: `Describe the revenue share terms for tenant ${tenantIndex}, lease ${leaseIndex}.`,
        input: 'textarea',
      };
    case 'extensions':
      return {
        id: baseId,
        prompt: `What extension options exist for tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'textarea',
        helpText: 'Include term length, notice windows, and any rent resets.',
      };
    case 'chargeables':
      return {
        id: baseId,
        prompt: `Which chargeable items apply to tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'textarea',
        helpText: 'For example CAM, taxes, insurance, utilities, parking, percentage rent, reimbursements.',
      };
    case 'special_terms':
      return {
        id: baseId,
        prompt: `What other special terms apply to tenant ${tenantIndex}, lease ${leaseIndex}?`,
        input: 'textarea',
        helpText: 'Include exclusivity, tenant improvement obligations, co-tenancy, termination rights, guarantees, or renewal quirks.',
      };
    default:
      return null as never;
  }
}

function nextLeaseStep(answers: ChatAnswers, tenantIndex: number, leaseIndex: number): string | null {
  const totalLeases = leaseCount(answers, tenantIndex);
  if (leaseIndex < totalLeases) {
    return `tenant_${tenantIndex}_lease_${leaseIndex + 1}_suite`;
  }
  const totalTenants = tenantCount(answers);
  if (tenantIndex < totalTenants) {
    return `tenant_${tenantIndex + 1}_name`;
  }
  return 'exit_assumptions';
}

export const commercialRealEstateScript: ChatScript = {
  id: 'commercial-real-estate',
  title: 'Commercial Real Estate Chat',
  completeLabel: 'Finish Chat',
  intro: 'Let’s capture the property structure, then walk tenant by tenant through lease details, extensions, revenue share, and chargeable items.',
  initialStepId: 'property_summary',
  resolveStep: (stepId: string, answers: ChatAnswers) => {
    if (stepId === 'property_summary') {
      return {
        id: stepId,
        prompt: 'What kind of property is this, and what is the high-level leasing story?',
        input: 'textarea',
      };
    }

    if (stepId === 'tenant_count') {
      return {
        id: stepId,
        prompt: 'How many tenants should we capture for this lot or property?',
        input: 'number',
      };
    }

    if (stepId === 'occupancy_context') {
      return {
        id: stepId,
        prompt: 'What occupancy, rollover, or vacancy context should we keep in mind before looking at individual tenants?',
        input: 'textarea',
      };
    }

    if (stepId === 'exit_assumptions') {
      return {
        id: stepId,
        prompt: 'What exit assumptions or biggest leasing risks should be tracked?',
        input: 'textarea',
        submitLabel: 'Finish Chat',
      };
    }

    const tenantMatch = stepId.match(/^tenant_(\d+)_name$/);
    if (tenantMatch) {
      return tenantNameStep(Number(tenantMatch[1]));
    }

    const leaseCountMatch = stepId.match(/^tenant_(\d+)_lease_count$/);
    if (leaseCountMatch) {
      return tenantLeaseCountStep(Number(leaseCountMatch[1]));
    }

    const leaseFieldMatch = stepId.match(/^tenant_(\d+)_lease_(\d+)_(suite|lease_type|start|end|term_notes|base_rent|rent_growth|revenue_share|revenue_share_terms|extensions|chargeables|special_terms)$/);
    if (leaseFieldMatch) {
      return leaseFieldStep(Number(leaseFieldMatch[1]), Number(leaseFieldMatch[2]), leaseFieldMatch[3]);
    }

    return null;
  },
  getNextStepId: (currentStepId: string, answer: string | number | boolean, answers: ChatAnswers) => {
    if (currentStepId === 'property_summary') {
      return 'tenant_count';
    }

    if (currentStepId === 'tenant_count') {
      return 'occupancy_context';
    }

    if (currentStepId === 'occupancy_context') {
      return tenantCount(answers) > 0 ? 'tenant_1_name' : 'exit_assumptions';
    }

    const tenantMatch = currentStepId.match(/^tenant_(\d+)_name$/);
    if (tenantMatch) {
      return `tenant_${tenantMatch[1]}_lease_count`;
    }

    const leaseCountMatch = currentStepId.match(/^tenant_(\d+)_lease_count$/);
    if (leaseCountMatch) {
      const tenantIndex = Number(leaseCountMatch[1]);
      return leaseCount({ ...answers, [currentStepId]: answer }, tenantIndex) > 0
        ? `tenant_${tenantIndex}_lease_1_suite`
        : nextLeaseStep({ ...answers, [currentStepId]: answer }, tenantIndex, 0);
    }

    const leaseFieldMatch = currentStepId.match(/^tenant_(\d+)_lease_(\d+)_(suite|lease_type|start|end|term_notes|base_rent|rent_growth|revenue_share|revenue_share_terms|extensions|chargeables|special_terms)$/);
    if (leaseFieldMatch) {
      const tenantIndex = Number(leaseFieldMatch[1]);
      const leaseIndex = Number(leaseFieldMatch[2]);
      const field = leaseFieldMatch[3];
      if (field === 'suite') return `tenant_${tenantIndex}_lease_${leaseIndex}_lease_type`;
      if (field === 'lease_type') return `tenant_${tenantIndex}_lease_${leaseIndex}_start`;
      if (field === 'start') return `tenant_${tenantIndex}_lease_${leaseIndex}_end`;
      if (field === 'end') return `tenant_${tenantIndex}_lease_${leaseIndex}_term_notes`;
      if (field === 'term_notes') return `tenant_${tenantIndex}_lease_${leaseIndex}_base_rent`;
      if (field === 'base_rent') return `tenant_${tenantIndex}_lease_${leaseIndex}_rent_growth`;
      if (field === 'rent_growth') return `tenant_${tenantIndex}_lease_${leaseIndex}_revenue_share`;
      if (field === 'revenue_share') {
        return answer ? `tenant_${tenantIndex}_lease_${leaseIndex}_revenue_share_terms` : `tenant_${tenantIndex}_lease_${leaseIndex}_extensions`;
      }
      if (field === 'revenue_share_terms') return `tenant_${tenantIndex}_lease_${leaseIndex}_extensions`;
      if (field === 'extensions') return `tenant_${tenantIndex}_lease_${leaseIndex}_chargeables`;
      if (field === 'chargeables') return `tenant_${tenantIndex}_lease_${leaseIndex}_special_terms`;
      if (field === 'special_terms') {
        return nextLeaseStep(answers, tenantIndex, leaseIndex);
      }
    }

    if (currentStepId === 'exit_assumptions') {
      return null;
    }

    return null;
  },
  buildSummary: (answers: ChatAnswers) => {
    const lines: string[] = [
      'Commercial Real Estate Chat Notes',
      `Property summary: ${answers.property_summary || '-'}`,
      `Tenant count: ${answers.tenant_count || 0}`,
      `Occupancy / rollover context: ${answers.occupancy_context || '-'}`,
    ];

    for (let tenantIndex = 1; tenantIndex <= tenantCount(answers); tenantIndex += 1) {
      const tenantLabel = answers[`tenant_${tenantIndex}_name`] || `Tenant ${tenantIndex}`;
      lines.push(`Tenant ${tenantIndex}: ${tenantLabel}`);
      lines.push(`  Lease count: ${answers[`tenant_${tenantIndex}_lease_count`] || 0}`);
      for (let leaseIndex = 1; leaseIndex <= leaseCount(answers, tenantIndex); leaseIndex += 1) {
        lines.push(`  Lease ${leaseIndex}`);
        lines.push(`    Space: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_suite`] || '-'}`);
        lines.push(
          `    Lease type: ${choiceLabel(
            leaseFieldStep(tenantIndex, leaseIndex, 'lease_type'),
            answers[`tenant_${tenantIndex}_lease_${leaseIndex}_lease_type`]
          )}`
        );
        lines.push(`    Start: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_start`] || '-'}`);
        lines.push(`    End: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_end`] || '-'}`);
        lines.push(`    Term details: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_term_notes`] || '-'}`);
        lines.push(`    Base rent: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_base_rent`] || '-'}`);
        lines.push(`    Rent growth: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_rent_growth`] || '-'}`);
        lines.push(`    Revenue share: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_revenue_share`] ? 'Yes' : 'No'}`);
        if (answers[`tenant_${tenantIndex}_lease_${leaseIndex}_revenue_share_terms`]) {
          lines.push(`    Revenue share terms: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_revenue_share_terms`]}`);
        }
        lines.push(`    Extensions: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_extensions`] || '-'}`);
        lines.push(`    Chargeables: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_chargeables`] || '-'}`);
        lines.push(`    Special terms: ${answers[`tenant_${tenantIndex}_lease_${leaseIndex}_special_terms`] || '-'}`);
      }
    }

    lines.push(`Exit assumptions / risks: ${answers.exit_assumptions || '-'}`);
    return lines.join('\n');
  },
};
