'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { TransformE164PhoneNumber } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    TransformE164PhoneNumber,
    PlainMessage<TransformE164PhoneNumber>
  > {}

export default function UserDefinedTransformE164NumberForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription className="w-[90%]">
            Set the length of the output e164 phone number to be the same as the
            input e164 phone number.
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(new TransformE164PhoneNumber({ preserveLength: checked }))
          }
          disabled={isDisabled}
        />
      </div>
      <FormErrorMessage message={errors?.preserveLength?.message} />
    </div>
  );
}
