pragma solidity ^0.4.24;

import { SafeMath } from "./SafeMath.sol";
import {RLP} from "./rlplib.sol";
import { RLPEncode } from "./rlpencode.sol";
import { BytesLib } from "./byteslib.sol";
import "./ecverify.sol";
contract ValidatorSet is ECVerify {
  using SafeMath for uint256;
  using SafeMath for uint8;
//   using BytesLib for bytes32;
  int256 constant INT256_MIN = -int256((2**255)-1);
  uint256 constant UINT256_MAX = (2**256)-1;
  using RLP for bytes;
  using RLP for RLP.RLPItem;
  using RLP for RLP.Iterator;
  event NewProposer(address indexed user, bytes data);

  struct Validator {
    uint256 votingPower;
    int256 accumulator;
    address validator;
    // bytes32 pubkey1;
    // bytes32 pubkey2;
    string pubkey;
  }

  address public proposer;
  uint256 public totalVotingPower;
  uint256 public lowestPower;
  Validator[] public validators;


  constructor() public {
    totalVotingPower = 0;
    lowestPower =  UINT256_MAX;
  }

  function addValidator(address validator, uint256 votingPower, string _pubkey) public {
    require(votingPower > 0);
    validators.push(Validator(votingPower, 0, validator,_pubkey)); //use index instead

    if (lowestPower > votingPower ) {
      lowestPower = votingPower;
    }

    totalVotingPower = totalVotingPower.add(votingPower);
  }
  function getPubkey(uint256 index) public view returns(string){
    //  return BytesLib.concat(abi.encodePacked(validators[index].pubkey1), abi.encodePacked(validators[index].pubkey2));
    return validators[index].pubkey;
  }

    function getValidatorSet() public view returns (uint256[] ,address[]){
        uint256[] memory powers= new uint256[](validators.length);
        address[] memory validatorAddresses= new address[](validators.length);
        for (uint8 i = 0; i < validators.length; i++) {
            validatorAddresses[i]=validators[i].validator;
            powers[i]=validators[i].votingPower;

        }

        return (powers,validatorAddresses);
    }


  function selectProposer() public returns(address) {
    require(validators.length > 0);

    for (uint8 i = 0; i < validators.length; i++) {
      validators[i].accumulator += int(validators[i].votingPower);
    }

    int256 max = INT256_MIN;
    uint8 index = 0;

    for (i = 0; i < validators.length; i++) {
      if (max < validators[i].accumulator){
        max = validators[i].accumulator;
        index = i;
      }
    }

    validators[index].accumulator -= int(totalVotingPower);
    proposer = validators[index].validator;

    emit NewProposer(proposer, "0");

    return proposer;
  }
   // Inputs: start,end,roothash,vote bytes,signatures of validators ,tx(extradata)
    // Constants: chainid,type,votetype
    // extradata => start,end ,proposer etc rlp encoded hash
    //
    //
    // todo : check proposer verify signatures  for validators
    bytes public chainID = "test-chain-5w6Ce4";
    bytes public roundType = "vote";
    bytes public voteType = "0x02";

    // constructor (bytes _chainID, bytes _rountType , bytes _voteType ){
    //     chainID=_chainID;
    //     roundType=_rountType;
    //     voteType=_voteType;
    // }

    // @params start-> startBlock , end-> EndBlock , roothash-> merkelRoot
    function validate(bytes vote,bytes sigs,bytes extradata)public view returns(bool,address) {
        // add require to check msg,sender == proposer
        RLP.RLPItem[] memory dataList = vote.toRLPItem().toList();
        require(keccak256(dataList[0].toData())==keccak256(chainID),"Chain ID not same");
        require(keccak256(dataList[1].toData())==keccak256(roundType),"Round Type Not Same ");

        // require(keccak256(dataList[5].toData())==keccak256(_voteType),"Vote Type Not Same");

        // validate extra data using getSha256(extradata)
        require(keccak256(dataList[4].toData())==keccak256(getSha256(extradata)));

        // decode extra data and validate start,end etc
        RLP.RLPItem[] memory txDataList;
        txDataList=extradata.toRLPItem().toList()[0].toList();

        // require(txDataList[1].toUint() == start,"Start Block Does Not Match");
        // require(txDataList[2].toUint() == end, "End Block Does Not Match");

        // validate with get proposer from stake manager
         require(txDataList[0].toAddress()==proposer,"Message Sender!=Proposer in extra data");

        //slice sigs and do ecrecover from validator set
        bytes32 hash = keccak256(vote);
        // TODO change this as sigs would be concat of all sigs
        address validator;
        validator = ecrecovery(hash,sigs);

        return (true,validator);

    }
    function setChainId(string _chainID) public {
        chainID = bytes(_chainID);
    }
    function getSha256(bytes input) public returns (bytes20) {
        bytes32 hash = sha256(input);
        return bytes20(hash);

    }
}